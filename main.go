package main

import (
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"

	"github.com/gocql/gocql"

	"github.com/DemianSV/chrdsclient"
)

//go:embed all:assets
var Assets embed.FS

//go:embed index.html
var Index []byte

const (
	// Version
	Version string = "1.0.5"
	// LogFile
	LogFile string = "chrdsuimanager.log"
	// ConfigFile
	ConfigFile string = "chrdsuimanager.json"
)

var ConsistencyRead gocql.Consistency

// FlagLog File logistics type - Stdout
var FlagLog string
var Session *gocql.Session
var TokenAuth *jwtauth.JWTAuth
var TokenSecret = randString(16)

type ResponseCountT struct {
	Count int `json:"count"`
}

type DBConnectObserverT struct {
	Host    string `json:"host"`
	DC      string `json:"dc"`
	Rack    string `json:"rack"`
	Version string `json:"version"`
	Stat    string `json:"stat"`
	Up      bool   `json:"up"`
}
type DBConnectObserverAT []DBConnectObserverT

var DBStatus DBConnectObserverT
var DBStatusA DBConnectObserverAT

func init() {
	flag.StringVar(&FlagLog, "log", "stdout", "[-log file, stdout (default)]")
	flag.Parse()
	if FlagLog == "file" {
		fileLog, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			fmt.Println("Ошибка открытия LOG файла!")
			os.Exit(-1)
		}
		log.SetOutput(fileLog)
	}

	err := AppConfig.LoadConfig()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации (", err.Error(), ")!")
		return
	}

	chrdsclient.Conf.SpaceID = AppConfig.UIMANAGER.SPACEID
	chrdsclient.Conf.ModuleID = AppConfig.UIMANAGER.MODULEID
	chrdsclient.Conf.DataManagerURL = AppConfig.UIMANAGER.DATAMANAGERURL
	chrdsclient.Conf.ClientInSecureSkipVerify = AppConfig.HTTP.CLIENTINSECURE
	chrdsclient.Conf.DataManagerTimeOut = 1
}

func done() {
}

func (o *DBConnectObserverT) ObserveConnect(observed gocql.ObservedConnect) {
	if observed.Err != nil {
		log.Printf(
			"Connect observed: host=%s start=%v end=%v err=%v",
			observed.Host.ConnectAddress(),
			observed.Start,
			observed.End,
			observed.Err,
		)
	} else {
		log.Printf(
			"Connect observed: host=%s start=%v end=%v",
			observed.Host.ConnectAddress(),
			observed.Start,
			observed.End,
		)
	}

	DBStatus.Host = observed.Host.ConnectAddress().String()
	DBStatus.DC = observed.Host.DataCenter()
	DBStatus.Rack = observed.Host.Rack()
	DBStatus.Version = observed.Host.Version().String()
	DBStatus.Stat = observed.Host.State().String()
	DBStatus.Up = observed.Host.IsUp()

	for i, item := range DBStatusA {
		if item.Host == DBStatus.Host {
			DBStatusA[i] = DBStatus
			return
		}
	}
	DBStatusA = append(DBStatusA, DBStatus)
}

func main() {
	defer done()
	defer log.Print("The service is stopped!")

	envFlag := runtime.GOMAXPROCS(runtime.NumCPU())
	if envFlag > -1 {
		log.Print("GOMAXPROCS = ", runtime.NumCPU())
	} else {
		log.Print("GOMAXPROCS is default!")
	}

	// Connect to DB cluster
	cluster := gocql.NewCluster(AppConfig.DB.HOSTS...)
	cluster.Timeout = time.Duration(AppConfig.DB.TIMEOUT) * time.Millisecond
	cluster.Keyspace = AppConfig.DB.KEYSPACE

	AppConfig.DB.CONSISTENCY = strings.ToUpper(AppConfig.DB.CONSISTENCY)
	switch AppConfig.DB.CONSISTENCY {
	case "QUORUM":
		cluster.Consistency = gocql.Quorum
	case "ANY":
		cluster.Consistency = gocql.Any
	case "ONE":
		cluster.Consistency = gocql.One
	case "TWO":
		cluster.Consistency = gocql.Two
	case "THREE":
		cluster.Consistency = gocql.Three
	case "ALL":
		cluster.Consistency = gocql.All
	case "LOCALQUORUM":
		cluster.Consistency = gocql.LocalQuorum
	case "EACHQUORUM":
		cluster.Consistency = gocql.EachQuorum
	case "LOCALONE":
		cluster.Consistency = gocql.LocalOne
	default:
		cluster.Consistency = gocql.Quorum
	}
	AppConfig.DB.CONSISTENCYREAD = strings.ToUpper(AppConfig.DB.CONSISTENCYREAD)
	switch AppConfig.DB.CONSISTENCYREAD {
	case "QUORUM":
		ConsistencyRead = gocql.Quorum
	case "ANY":
		ConsistencyRead = gocql.Any
	case "ONE":
		ConsistencyRead = gocql.One
	case "TWO":
		ConsistencyRead = gocql.Two
	case "THREE":
		ConsistencyRead = gocql.Three
	case "ALL":
		ConsistencyRead = gocql.All
	case "LOCALQUORUM":
		ConsistencyRead = gocql.LocalQuorum
	case "EACHQUORUM":
		ConsistencyRead = gocql.EachQuorum
	case "LOCALONE":
		ConsistencyRead = gocql.LocalOne
	default:
		ConsistencyRead = gocql.Quorum
	}

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: AppConfig.DB.USERNAME,
		Password: AppConfig.DB.PASSWORD,
	}

	if AppConfig.DB.TLS {
		cluster.SslOpts = &gocql.SslOptions{
			EnableHostVerification: AppConfig.DB.HOSTVERIFICATION,
			CertPath:               AppConfig.DB.CERTPATH,
			KeyPath:                AppConfig.DB.KEYPATH,
			CaPath:                 AppConfig.DB.CAPATH,
		}
	}

	cluster.ConnectObserver = &DBConnectObserverT{}

	var err error
	Session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal("It was not possible to establish a connection with the database (", err, ")!")
	} else {
		log.Print("The connection to the database is established... OK")
	}
	defer Session.Close()

	r := chi.NewRouter() // Initialization Gochi Router

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger) // Expanded logging
	r.Use(middleware.URLFormat)

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  allowOriginFunc,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(httprate.Limit(
		AppConfig.HTTP.RL, // Requests
		1*time.Second,     // Per duration
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			return r.Header.Get("api_key"), nil
		}),
	))

	r.Group(func(r chi.Router) {
		TokenAuth = jwtauth.New("HS256", []byte(TokenSecret), nil)
		r.Use(jwtauth.Verifier(TokenAuth))
		r.Use(jwtauth.Authenticator(TokenAuth))

		r.Route("/api/v{version}/version", func(r chi.Router) {
			r.Get("/", getVersion) // Version
		})
		r.Route("/api/v{version}/userinfo", func(r chi.Router) {
			r.Get("/", getUserInfo) // Information on an authorized user
		})
		r.Route("/api/v{version}/password", func(r chi.Router) {
			r.Put("/", putPassword) // Password change
		})
	})

	// Administrator branch
	r.Route("/api/v{version}/admin", func(r chi.Router) {
		TokenAuth = jwtauth.New("HS256", []byte(TokenSecret), nil)
		r.Use(jwtauth.Verifier(TokenAuth))
		r.Use(jwtauth.Authenticator(TokenAuth))

		r.Route("/user", func(r chi.Router) {
			r.Get("/count", getUserCount)
			r.Get("/select", getUserSelect)
			r.Put("/update", putUserUpdate)
			r.Put("/remove", putUserRemove)
			r.Put("/create", putUserCreate)
		})

		r.Route("/registration", func(r chi.Router) {
			r.Get("/count", getModuleCount)
			r.Get("/select", getModuleSelect)
			r.Put("/update", putModuleUpdate)
			r.Put("/remove", putModuleRemove)
			r.Put("/create", putModuleCreate)
		})

		r.Route("/task", func(r chi.Router) {
			r.Get("/count", getTaskCount)
			r.Get("/select", getTaskSelect)
			r.Put("/update", putTaskUpdate)
			r.Put("/remove", putTaskRemove)
			r.Put("/create", putTaskCreate)
		})

		r.Route("/space", func(r chi.Router) {
			r.Get("/count", getSpaceCount)
			r.Get("/select", getSpaceSelect)
			r.Put("/update", putSpaceUpdate)
			r.Put("/remove", putSpaceRemove)
			r.Put("/create", putSpaceCreate)
		})

		r.Route("/problem", func(r chi.Router) {
			r.Get("/count", getProblemCount)
			r.Get("/select", getProblemSelect)
			// r.Put("/update", putProblemUpdate)
		})

		r.Route("/rawdata", func(r chi.Router) {
			r.Get("/count", getRawDataCount)
			r.Put("/select", putRawDataSelect)
			r.Put("/metric/select", putMetricDataSelect)
		})

		r.Route("/rawtext", func(r chi.Router) {
			r.Get("/count", getRawTextCount)
			r.Put("/select", putRawTextSelect)
			r.Put("/metric/select", putMetricTextSelect)
		})

		r.Route("/chartdata", func(r chi.Router) {
			r.Put("/", putChartData)
		})

		r.Route("/database", func(r chi.Router) {
			r.Get("/status", getDataBaseStatus)
		})

		r.Route("/datamanager", func(r chi.Router) {
			r.Get("/status", getDataManagerStatus)
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/count", getDashboardCount)
			r.Get("/select", getDashboardSelect)
			r.Put("/data", putChartDashboardData)
			r.Put("/create", putDashboardCreate)
			r.Put("/remove", putDashboardRemove)
			r.Put("/edit", putDashboardEdit)
		})

		r.Route("/emaillist", func(r chi.Router) {
			r.Get("/select", getEMailListSelect)
			r.Put("/create", putEMailListCreate)
			r.Put("/remove", putEMailListRemove)
			r.Put("/edit", putEMailListEdit)
		})
	})

	r.Route("/api/v{version}/exit", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			cookie := &http.Cookie{
				Name:     "jwt",
				MaxAge:   -1,
				HttpOnly: true,
				Path:     "/",
			}
			http.SetCookie(w, cookie)
			w.WriteHeader(http.StatusOK)
			log.Print("Remove cookie!")
		})
	})

	// Requests to UI
	r.Get("/", webTmpl)
	r.Get("/user", webTmpl)
	r.Get("/registration", webTmpl)
	r.Get("/task", webTmpl)
	r.Get("/space", webTmpl)
	r.Get("/about", webTmpl)
	r.Get("/rawtext", webTmpl)
	r.Get("/rawdata", webTmpl)
	r.Get("/dashboard", webTmpl)
	r.Get("/database", webTmpl)
	r.Get("/metric", webTmpl)
	r.Get("/problem", webTmpl)
	r.Get("/settings", webTmpl)

	// Authorization request
	r.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		log.Print(ua)
		if token, err := webAuth(r.FormValue("username"), r.FormValue("password")); err == nil && string(token) != "" {
			cookie := &http.Cookie{
				Name:     "jwt",
				Value:    string(token),
				MaxAge:   21600,
				HttpOnly: true,
				Path:     "/",
			}
			http.SetCookie(w, cookie)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	// Static content for UI
	fs, errFS := fs.Sub(Assets, "assets")
	if errFS != nil {
		log.Fatal("It was not possible to connect static content!")
	} else {
		fileServer(r, "/assets", http.FS(fs))
	}

	if AppConfig.HTTP.TLS { // HTTPS connection
		crt, err := tls.LoadX509KeyPair(AppConfig.HTTP.CERTPATH, AppConfig.HTTP.KEYPATH)
		if err != nil {
			log.Fatal("Failed to download the certificate (", err, ")!")
			return
		}
		var crtPool *x509.CertPool
		if AppConfig.HTTP.CAPATH == "" {
			crtPool = nil
		} else {
			crtPool = x509.NewCertPool()
			if crtCA, err := os.ReadFile(AppConfig.HTTP.CAPATH); err != nil {
				log.Fatal("Failed to download CA certificate (", err, ")!")
			} else if ok := crtPool.AppendCertsFromPEM(crtCA); !ok {
				log.Fatal("It was not possible to apply CA certificate!")
			}
		}
		tlsConfig := &tls.Config{
			RootCAs:      crtPool,
			Certificates: []tls.Certificate{crt},
		}

		httpServer := &http.Server{Addr: AppConfig.HTTP.HOST + ":" + AppConfig.HTTP.PORT,
			Handler:      r,
			ReadTimeout:  time.Duration(AppConfig.HTTP.READTIMEOUT) * time.Second,
			WriteTimeout: time.Duration(AppConfig.HTTP.WRITETIMEOUT) * time.Second,
			TLSConfig:    tlsConfig}
		log.Print("Сервис UIManager запущен на порту ", AppConfig.HTTP.PORT, "(HTTPS)... ОК")
		httpServer.ListenAndServeTLS("", "")
	} else { // HTTP connection
		httpServer := &http.Server{Addr: AppConfig.HTTP.HOST + ":" + AppConfig.HTTP.PORT,
			Handler:      r,
			ReadTimeout:  time.Duration(AppConfig.HTTP.READTIMEOUT) * time.Second,
			WriteTimeout: time.Duration(AppConfig.HTTP.WRITETIMEOUT) * time.Second}
		log.Print("Сервис UIManager запущен на порту ", AppConfig.HTTP.PORT, "(HTTP)... ОК")
		httpServer.ListenAndServe()
	}
}

func webTmpl(w http.ResponseWriter, r *http.Request) {
	uaClient := r.Header.Get("User-Agent")
	raClient := r.Header.Get("X-Real-Ip")
	if raClient == "" {
		raClient = r.Header.Get("X-Forwarded-For")
	}
	if raClient == "" {
		raClient = r.RemoteAddr
	}
	if raClient == "" {
		raClient = "Address not determined"
	}
	log.Print("IP: ", raClient, " UserAgent: ", uaClient)

	w.Header().Set("Cache-control", "no-cache, must-revalidate, private, no-store, s-maxage=0, max-age=0, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "-10d")

	w.Write(Index)
}

func getVersion(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the API version!", claims["username"])

	type VersionJSON struct {
		Name    string
		Version string
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		version := VersionJSON{
			Name:    "Χάρυβδις UIManager",
			Version: Version,
		}
		jsonAPI, err := json.Marshal(version)
		if err != nil {
			log.Print("JSON ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonAPI)
	}
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func webAuth(userName string, password string) ([]byte, error) {
	ctx := context.Background()
	scanner := Session.Query(`SELECT user_id, status, password, role, owner_id FROM user WHERE user_name = ?`, userName).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
	for scanner.Next() {
		var status int
		var userID string
		var passwordDB string
		var role string
		var ownerID string
		err := scanner.Scan(&userID, &status, &passwordDB, &role, &ownerID)
		if err != nil {
			log.Print(err)
			return []byte(""), err
		} else {
			if status == 1 && passwordDB == password {
				TokenAuth = jwtauth.New("HS256", []byte(TokenSecret), nil)
				claims := jwt.MapClaims{"username": userName, "userid": userID, "role": role, "ownerid": ownerID}
				jwtauth.SetExpiryIn(claims, 6*time.Hour) // Life life of token 6 hours
				jwtauth.SetIssuedNow(claims)             // The time to create token
				_, tokenString, err := TokenAuth.Encode(claims)
				if err != nil {
					log.Print("User " + userName + " ACCESS DENIED!")
					{
						err := chrdsclient.Log("audit", "User "+userName+" ACCESS DENIED")
						if err != nil {
							log.Print("It was not possible to record the audit data (", err, ")!")
						}
					}
					log.Print("Access error (", err.Error(), ")!")
					return []byte(""), err
				}
				if tokenString != "" {
					log.Print("The user " + userName + " has entered the system!")
					{
						err := chrdsclient.Log("audit", "User "+userName+" entered the system")
						if err != nil {
							log.Print("It was not possible to record the audit data (", err, ")!")
							log.Print("Access error (", err.Error(), ")!")
							return []byte(""), err
						}
					}

					err := Session.Query(`UPDATE user SET login_time = toTimestamp(now()) WHERE user_name = ?`, userName).WithContext(ctx).Exec()
					if err != nil {
						log.Print("Failed to update the date of the user entry ", userName, "!")
						go chrdsclient.Log("log", "Failed to update the date of the user entry "+userName)
					}

					return []byte(tokenString), nil
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return []byte(""), err
	}

	err := errors.New("ACCESS DENIED")
	log.Print("User " + userName + " ACCESS DENIED!")
	go chrdsclient.Log("audit", "User "+userName+" ACCESS DENIED")
	return []byte(""), err
}

func randString(n int) string {
	var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	l := len(letters)
	b := make([]byte, n)
	crand.Read(b)
	for i := range n {
		b[i] = letters[int(b[i])%l]
	}
	return string(b)
}

func allowOriginFunc(r *http.Request, origin string) bool {
	return false
}

func getRawDataCount(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT count(*) FROM raw_data01`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						spaceList = append(spaceList, spaceID)
					}
				}

				args := []interface{}{}
				for _, item := range spaceList {
					args = append(args, item)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM raw_data01 WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`) AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`,
						args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseCount.Count = 0
					responseCountJSON, err := json.Marshal(responseCount)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseCountJSON)
						return
					}
				}
			default:
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						spaceList = append(spaceList, spaceID)
					}
				}

				args := []interface{}{}
				for _, item := range spaceList {
					args = append(args, item)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM raw_data01 WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`) AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`,
						args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseCount.Count = 0
					responseCountJSON, err := json.Marshal(responseCount)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseCountJSON)
						return
					}
				}
			}
			for scanner.Next() {
				var count int = 0
				err := scanner.Scan(&count)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseCount.Count = count
					responseCountJSON, err := json.Marshal(responseCount)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseCountJSON)
						return
					}
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func getRawTextCount(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT count(*) FROM raw_text01`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						spaceList = append(spaceList, spaceID)
					}
				}

				args := []interface{}{}
				for _, item := range spaceList {
					args = append(args, item)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM raw_text01 WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`) AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseCount.Count = 0
					responseCountJSON, err := json.Marshal(responseCount)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseCountJSON)
						return
					}
				}
			default:
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						spaceList = append(spaceList, spaceID)
					}
				}

				args := []interface{}{}
				for _, item := range spaceList {
					args = append(args, item)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM raw_text01 WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`) AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseCount.Count = 0
					responseCountJSON, err := json.Marshal(responseCount)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseCountJSON)
						return
					}
				}
			}
			for scanner.Next() {
				var count int = 0
				err := scanner.Scan(&count)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseCount.Count = count
					responseCountJSON, err := json.Marshal(responseCount)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseCountJSON)
						return
					}
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putRawTextSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested raw_text data!", claims["username"])
	go chrdsclient.Log("log", fmt.Sprintf("User %v requested raw_text data", claims["username"]))

	type ResponseRawTextPageT struct {
		DateMin    int64 `json:"datemin"`
		DateMax    int64 `json:"datemax"`
		DataCount  int64 `json:"datacount"`
		PageCount  int64 `json:"pagecount"`
		PageCurent int64 `json:"pagecurent"`
	}
	type ResponseRawTextDataT struct {
		SpaceID    string `json:"spaceid"`
		Metric     string `json:"metric"`
		CreateTime int64  `json:"createtime"`
		EventTime  int64  `json:"eventtime"`
		Status     int    `json:"status"`
		Value      string `json:"value"`
		Object     string `json:"object"`
		SpaceDesc  string `json:"spacedesc"`
	}
	type ResponseRawTextDataAT []ResponseRawTextDataT
	type ResponseRawTextT struct {
		Data ResponseRawTextDataAT `json:"data"`
		Page ResponseRawTextPageT  `json:"page"`
	}

	type RequestT struct {
		SpaceID    string `json:"spaceid"`
		Metric     string `json:"metric"`
		PageCurent int64  `json:"pagecurent"`
	}

	var spaceDesc string

	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var request RequestT
			if err := json.Unmarshal(b, &request); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var responseRawTextData ResponseRawTextDataT
			var responseRawTextDataA ResponseRawTextDataAT
			var responseRawTextPage ResponseRawTextPageT

			ctx := context.Background()
			var scanner gocql.Scanner
			{
				scannerSpace := Session.Query(`SELECT description FROM space WHERE id = ?`, request.SpaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scannerSpace.Next() {
					err := scannerSpace.Scan(&spaceDesc)
					if err != nil {
						log.Print(err)
					}
				}

				args := []interface{}{}
				args = append(args, request.SpaceID)
				if request.Metric != "" {
					args = append(args, request.Metric)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if request.Metric != "" {
					scanner = Session.Query(`SELECT count(*), min(event_time), max(event_time) FROM raw_text02 WHERE space_id = ? AND metric = ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					scanner = Session.Query(`SELECT count(*), min(event_time), max(event_time) FROM raw_text01 WHERE space_id = ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				}
				for scanner.Next() {
					var dataCount int64
					var dateMin int64
					var dateMax int64

					err := scanner.Scan(&dataCount, &dateMin, &dateMax)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						return
					} else {
						responseRawTextPage.DataCount = dataCount

						var pageSize int64 = 24 * 60 * 60 * 1000
						pageCount := int64((dateMax - dateMin) / pageSize)
						if pageCount < 2 {
							responseRawTextPage.PageCount = 1
							responseRawTextPage.PageCurent = 1
							responseRawTextPage.DateMin = dateMin
							responseRawTextPage.DateMax = dateMax
						} else {
							responseRawTextPage.PageCount = pageCount
							responseRawTextPage.PageCurent = request.PageCurent
							responseRawTextPage.DateMin = dateMax - (pageSize * request.PageCurent) + 1
							responseRawTextPage.DateMax = dateMax - (pageSize * request.PageCurent) + pageSize
							if pageCount == request.PageCurent {
								responseRawTextPage.DateMin = dateMin
							}
						}

					}
				}

				if err := scanner.Err(); err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
					return
				}
			}

			{
				args := []interface{}{}

				if request.Metric != "" {
					args = append(args, request.SpaceID, request.Metric, responseRawTextPage.DateMin, responseRawTextPage.DateMax)
				} else {
					args = append(args, request.SpaceID, responseRawTextPage.DateMin, responseRawTextPage.DateMax)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if request.Metric != "" {
					scanner = Session.Query(`SELECT space_id, metric, create_time, event_time, status, value, object FROM raw_text02 WHERE space_id = ? AND metric = ? AND event_time >= ? AND event_time <= ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					scanner = Session.Query(`SELECT space_id, metric, create_time, event_time, status, value, object FROM raw_text01 WHERE space_id = ? AND event_time >= ? AND event_time <= ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				}
				for scanner.Next() {
					var spaceID string
					var metric string
					var createTime int64
					var eventTime int64
					var status int
					var value string
					var object string

					err := scanner.Scan(&spaceID, &metric, &createTime, &eventTime, &status, &value, &object)
					if err != nil {
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						responseRawTextData.SpaceID = spaceID
						responseRawTextData.Metric = metric
						responseRawTextData.CreateTime = createTime
						responseRawTextData.EventTime = eventTime
						responseRawTextData.Status = status
						responseRawTextData.Value = value
						responseRawTextData.Object = object
						responseRawTextData.SpaceDesc = spaceDesc

						responseRawTextDataA = append(responseRawTextDataA, responseRawTextData)
					}
				}

				if err := scanner.Err(); err != nil {
					log.Print(err)
					go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			var responseRawText ResponseRawTextT
			responseRawText.Data = responseRawTextDataA
			responseRawText.Page = responseRawTextPage
			responseJSON, err := json.Marshal(responseRawText)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				go chrdsclient.Metric("httpstatus", float32(http.StatusOK))
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
		} else {
			log.Print("Failed to choose the data!")
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putRawDataSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested raw_data data!", claims["username"])
	go chrdsclient.Log("log", fmt.Sprintf("User %v requested raw_data data", claims["username"]))

	type ResponseRawDataPageT struct {
		DateMin    int64 `json:"datemin"`
		DateMax    int64 `json:"datemax"`
		DataCount  int64 `json:"datacount"`
		PageCount  int64 `json:"pagecount"`
		PageCurent int64 `json:"pagecurent"`
	}
	type ResponseRawDataDataT struct {
		SpaceID    string  `json:"spaceid"`
		Metric     string  `json:"metric"`
		CreateTime int64   `json:"createtime"`
		EventTime  int64   `json:"eventtime"`
		Status     int     `json:"status"`
		Value      float32 `json:"value"`
		Object     string  `json:"object"`
		SpaceDesc  string  `json:"spacedesc"`
	}
	type ResponseRawDataDataAT []ResponseRawDataDataT
	type ResponseRawDataT struct {
		Data ResponseRawDataDataAT `json:"data"`
		Page ResponseRawDataPageT  `json:"page"`
	}

	type RequestT struct {
		SpaceID    string `json:"spaceid"`
		Metric     string `json:"metric"`
		PageCurent int64  `json:"pagecurent"`
	}

	var spaceDesc string

	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var request RequestT
			if err := json.Unmarshal(b, &request); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var responseRawDataData ResponseRawDataDataT
			var responseRawDataDataA ResponseRawDataDataAT
			var responseRawDataPage ResponseRawDataPageT

			ctx := context.Background()
			var scanner gocql.Scanner
			{
				scannerSpace := Session.Query(`SELECT description FROM space WHERE id = ?`, request.SpaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scannerSpace.Next() {
					err := scannerSpace.Scan(&spaceDesc)
					if err != nil {
						log.Print(err)
					}
				}

				args := []interface{}{}
				args = append(args, request.SpaceID)
				if request.Metric != "" {
					args = append(args, request.Metric)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if request.Metric != "" {
					scanner = Session.Query(`SELECT count(*), min(event_time), max(event_time) FROM raw_data02 WHERE space_id = ? AND metric = ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					scanner = Session.Query(`SELECT count(*), min(event_time), max(event_time) FROM raw_data01 WHERE space_id = ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				}
				for scanner.Next() {
					var dataCount int64
					var dateMin int64
					var dateMax int64

					err := scanner.Scan(&dataCount, &dateMin, &dateMax)
					if err != nil {
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						responseRawDataPage.DataCount = dataCount

						var pageSize int64 = 24 * 60 * 60 * 1000
						pageCount := int64((dateMax - dateMin) / pageSize)
						if pageCount < 2 {
							responseRawDataPage.PageCount = 1
							responseRawDataPage.PageCurent = 1
							responseRawDataPage.DateMin = dateMin
							responseRawDataPage.DateMax = dateMax
						} else {
							responseRawDataPage.PageCount = pageCount
							responseRawDataPage.PageCurent = request.PageCurent
							responseRawDataPage.DateMin = dateMax - (pageSize * request.PageCurent) + 1
							responseRawDataPage.DateMax = dateMax - (pageSize * request.PageCurent) + pageSize
							if pageCount == request.PageCurent {
								responseRawDataPage.DateMin = dateMin
							}
						}

					}
				}

				if err := scanner.Err(); err != nil {
					log.Print(err)
					go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

			}

			{
				args := []interface{}{}
				if request.Metric != "" {
					args = append(args, request.SpaceID, request.Metric, responseRawDataPage.DateMin, responseRawDataPage.DateMax)
				} else {
					args = append(args, request.SpaceID, responseRawDataPage.DateMin, responseRawDataPage.DateMax)
				}
				syntKeyList := makeDateList(1)
				for _, item := range syntKeyList {
					args = append(args, item)
				}

				if request.Metric != "" {
					scanner = Session.Query(`SELECT space_id, metric, create_time, event_time, status, value, object FROM raw_data02 WHERE space_id = ? AND metric = ? AND event_time >= ? AND event_time <= ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					scanner = Session.Query(`SELECT space_id, metric, create_time, event_time, status, value, object FROM raw_data01 WHERE space_id = ? AND event_time >= ? AND event_time <= ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				}
				for scanner.Next() {
					var spaceID string
					var metric string
					var createTime int64
					var eventTime int64
					var status int
					var value float32
					var object string

					err := scanner.Scan(&spaceID, &metric, &createTime, &eventTime, &status, &value, &object)
					if err != nil {
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						responseRawDataData.SpaceID = spaceID
						responseRawDataData.Metric = metric
						responseRawDataData.CreateTime = createTime
						responseRawDataData.EventTime = eventTime
						responseRawDataData.Status = status
						responseRawDataData.Value = value
						responseRawDataData.Object = object
						responseRawDataData.SpaceDesc = spaceDesc

						responseRawDataDataA = append(responseRawDataDataA, responseRawDataData)
					}
				}

				if err := scanner.Err(); err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
					return
				}
			}
			var responseRawData ResponseRawDataT
			responseRawData.Data = responseRawDataDataA
			responseRawData.Page = responseRawDataPage
			responseJSON, err := json.Marshal(responseRawData)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				go chrdsclient.Metric("httpstatus", float32(http.StatusOK))
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
		} else {
			log.Print("Failed to choose the data!")
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putMetricTextSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a meter list!", claims["username"])

	type MetricT struct {
		Metric string `json:"metric"`
	}
	type MetricTA []MetricT

	type SpaceT struct {
		SpaceID string `json:"spaceid"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var space SpaceT
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var responseMetric MetricT
			var responseMetricA MetricTA
			ctx := context.Background()
			scanner := Session.Query(`SELECT metric FROM space_metric_text WHERE space_id = ?`, space.SpaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			for scanner.Next() {
				var metric string

				err := scanner.Scan(&metric)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseMetric.Metric = metric

					responseMetricA = append(responseMetricA, responseMetric)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseJSON, err := json.Marshal(responseMetricA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			log.Print("It was not possible to choose metrics!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putMetricDataSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a meter list!", claims["username"])

	type MetricT struct {
		Metric string `json:"metric"`
	}
	type MetricTA []MetricT

	type SpaceT struct {
		SpaceID string `json:"spaceid"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var space SpaceT
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var responseMetric MetricT
			var responseMetricA MetricTA
			ctx := context.Background()
			scanner := Session.Query(`SELECT metric FROM space_metric_data WHERE space_id = ?`, space.SpaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			for scanner.Next() {
				var metric string

				err := scanner.Scan(&metric)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseMetric.Metric = metric

					responseMetricA = append(responseMetricA, responseMetric)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseJSON, err := json.Marshal(responseMetricA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			log.Print("It was not possible to choose metrics!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func makeDateList(monthsNum int) []string {
	if monthsNum == 1 {
		monthsNum = 2
	}
	t := time.Now()
	var tStringA []string
	for i := 0; i <= monthsNum-1; i++ {
		tString := t.AddDate(0, -i, 0).Format("2006.01")
		tStringA = append(tStringA, tString)
	}
	return tStringA
}
