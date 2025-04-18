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
	rand "math/rand"
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
	Version string = "1.0.4"
	// Version string = "1.0.5"
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

type ResponsRegistrationT struct {
	ID          string `json:"id"`
	TypeMod     string `json:"type"`
	Status      int    `json:"status"`
	Description string `json:"description"`
	UserID      string `json:"userid"`
}
type ResponseRegistrationAT []ResponsRegistrationT

type ResponseTaskT struct {
	ID         string `json:"id"`
	ModuleID   string `json:"moduleid"`
	SpaceID    string `json:"spaceid"`
	Object     string `json:"object"`
	Metric     string `json:"metric"`
	Status     int    `json:"status"`
	Critical   string `json:"critical"`
	Warning    string `json:"warning"`
	Interval   int64  `json:"interval"`
	DataType   string `json:"datatype"`
	ModuleDesc string `json:"moduledesc"`
	SpaceDesc  string `json:"spacedesc"`
}
type ResponseTaskAT []ResponseTaskT

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
			r.Get("/count", getRegistrationCount)
			r.Get("/select", getRegistrationSelect)
			r.Put("/update", putRegistrationUpdate)
			r.Put("/remove", putRegistrationRemove)
			r.Put("/create", putRegistrationCreate)
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
			r.Get("/status", getDatabaseStatus)
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
	for i := 0; i < n; i++ {
		b[i] = letters[int(b[i])%l]
	}
	return string(b)
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested API userinfo!", claims["username"])

	type UserInfo struct {
		UserName string `json:"username"`
		FName    string `json:"fname"`
		LName    string `json:"lname"`
		UserID   string `json:"userid"`
		Role     string `json:"role"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		ctx := context.Background()
		scanner := Session.Query(`SELECT user_name, first_name, last_name, status, user_id, role FROM user WHERE user_name = ?`, claims["username"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
		for scanner.Next() {
			var userName string
			var firstName string
			var lastName string
			var status int
			var userID string
			var role string
			err := scanner.Scan(&userName, &firstName, &lastName, &status, &userID, &role)
			if err != nil {
				log.Print("ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				if status == 1 {
					userinfo := UserInfo{
						UserName: userName,
						FName:    firstName,
						LName:    lastName,
						UserID:   userID,
						Role:     role,
					}
					jsonAPI, err := json.Marshal(userinfo)
					if err != nil {
						log.Print("JSON ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.Write(jsonAPI)
				} else {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Print("ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func putPassword(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a password change!", claims["username"])

	type Passwords struct {
		CurPassword string `json:"curpassword"`
		NewPassword string `json:"newpassword"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var passwords Passwords
		if err := json.Unmarshal(b, &passwords); err != nil {
			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		{
			ctx := context.Background()
			scanner := Session.Query(`SELECT status, password FROM user WHERE user_name = ? LIMIT 1`, claims["username"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			for scanner.Next() {
				var status int
				var passwordDB string
				err := scanner.Scan(&status, &passwordDB)
				if err != nil {
					log.Print("Failed to change the password (SELECT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusNotFound)
					return
				} else {
					if passwordDB == passwords.CurPassword && status == 1 {
						err := Session.Query(`UPDATE user SET password = ? WHERE user_name = ?`, passwords.NewPassword, claims["username"]).WithContext(ctx).Exec()
						if err != nil {
							log.Print("Failed to change the password (UPDATE: ", err.Error(), ")!")
							w.WriteHeader(http.StatusInternalServerError)
							return
						} else {
							log.Printf("User %v successfully changed the password!", claims["username"])
							w.WriteHeader(http.StatusOK)
							return
						}
					}
				}
			}
		}
	}
	log.Print("It was not possible to change the user profile!")
	w.WriteHeader(http.StatusNotFound)
}

func putUserCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of the profile!", claims["username"])

	type User struct {
		UserName    string `json:"username"`
		Password    string `json:"password"`
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		Status      int    `json:"status"`
		Description string `json:"description"`
		Role        string `json:"role"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var user User
			if err := json.Unmarshal(b, &user); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				userID, err := gocql.RandomUUID()
				if err != nil {
					log.Print("Failed to create an UUID user (", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if claims["role"] == "admin" {
					user.Role = "user"
				}
				ctx := context.Background()
				err = Session.Query(`
				BEGIN BATCH
					INSERT INTO user (user_id, user_name, password, first_name, last_name, status, description, role, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
					INSERT INTO user_id (user_id, user_name) VALUES (?, ?)
					INSERT INTO owner_id_user (owner_id, user_name) VALUES (?, ?)
				APPLY BATCH
			`, userID, user.UserName, user.Password, user.FirstName, user.LastName, user.Status, user.Description, user.Role, claims["userid"], userID, user.UserName, claims["userid"], user.UserName).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a user profile (INSERT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				log.Printf("User %v successfully created a profile for the user %v!", claims["username"], user.UserName)
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			log.Print("It was not possible to create a profile!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putUserUpdate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a profile update!", claims["username"])

	type User struct {
		UserName    string `json:"username"`
		Password    string `json:"password"`
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		Status      int    `json:"status"`
		Description string `json:"description"`
		Role        string `json:"role"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var user User
			if err := json.Unmarshal(b, &user); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if claims["role"] == "admin" {
				user.Role = "user"
			}

			{
				ctx := context.Background()
				var err error
				if user.Password == "" {
					err = Session.Query(`UPDATE user SET first_name = ?, last_name = ?, status = ?, description = ?, role = ? WHERE user_name = ?`, user.FirstName, user.LastName, user.Status, user.Description, user.Role, user.UserName).WithContext(ctx).Exec()
				} else {
					err = Session.Query(`UPDATE user SET first_name = ?, last_name = ?, status = ?, description = ?, role = ?, password = ? WHERE user_name = ?`, user.FirstName, user.LastName, user.Status, user.Description, user.Role, user.Password, user.UserName).WithContext(ctx).Exec()
				}
				if err != nil {
					log.Print("Failed to change the user data (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully changed the user data %v!", claims["username"], user.UserName)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putUserRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a profile removal!", claims["username"])

	type User struct {
		UserName string `json:"username"`
		UserID   string `json:"userid"`
		OwnerID  string `json:"ownerid"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var user User
			if err := json.Unmarshal(b, &user); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
				BEGIN BATCH
					DELETE FROM user WHERE user_name = ?
					DELETE FROM user_id WHERE user_id = ?
					DELETE FROM owner_id_user WHERE owner_id = ? AND user_name = ?
				APPLY BATCH
				`, user.UserName, user.UserID, user.OwnerID, user.UserName).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to delete user data (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				log.Printf("User %v successfully deleted user profile %v!", claims["username"], user.UserName)
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			log.Print("It was not possible to remove the profile!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func allowOriginFunc(r *http.Request, origin string) bool {
	return false
}

func getUserCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		var userList []string

		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "admin" {
				scanner = Session.Query(`SELECT user_name FROM owner_id_user WHERE owner_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var userName string
					err := scanner.Scan(&userName)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						userList = append(userList, userName)
					}
				}
			}

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM user`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else {
				args := []interface{}{}
				for _, item := range userList {
					args = append(args, item)
				}
				if len(userList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM user WHERE user_name IN (?`+strings.Repeat(", ?", len(userList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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

func getSpaceCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM space`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
				scanner = Session.Query(`SELECT count(*) FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else {
				scanner = Session.Query(`SELECT count(*) FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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

func getTaskCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM task`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
			} else {
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
					log.Print(err)
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

func getRegistrationCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM registration`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
				scanner = Session.Query(`SELECT count(*) FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else {
				scanner = Session.Query(`SELECT count(*) FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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

func getUserSelect(w http.ResponseWriter, r *http.Request) {
	type ResponsUserT struct {
		UserID      string `json:"userid"`
		UserName    string `json:"username"`
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		Status      int    `json:"status"`
		LoginTime   int64  `json:"logintime"`
		Description string `json:"description"`
		Role        string `json:"role"`
		OwnerID     string `json:"ownerid"`
	}
	type ResponseUserAT []ResponsUserT
	var userList []string

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseUser ResponsUserT
			var responseUserA ResponseUserAT

			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "admin" {
				scanner = Session.Query(`SELECT user_name FROM owner_id_user WHERE owner_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var userName string
					err := scanner.Scan(&userName)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						userList = append(userList, userName)
					}
				}

				args := []interface{}{}
				for _, item := range userList {
					args = append(args, item)
				}
				if len(userList) > 0 {
					scanner = Session.Query(`SELECT user_name, description, first_name, last_name, login_time, status, user_id, role, owner_id FROM user WHERE user_name IN (?`+strings.Repeat(", ?", len(userList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseUserJSON, err := json.Marshal(responseUserA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseUserJSON)
						return
					}
				}
			} else {
				scanner = Session.Query(`SELECT user_name, description, first_name, last_name, login_time, status, user_id, role, owner_id FROM user`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			}

			for scanner.Next() {
				var userName string
				var description string
				var firstName string
				var lastName string
				var loginTime int64
				var status int
				var userID string
				var role string
				var ownerid string

				err := scanner.Scan(&userName, &description, &firstName, &lastName, &loginTime, &status, &userID, &role, &ownerid)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseUser.UserID = userID
					responseUser.UserName = userName
					responseUser.FirstName = firstName
					responseUser.LastName = lastName
					responseUser.LoginTime = loginTime
					responseUser.Status = status
					responseUser.Description = description
					responseUser.Role = role
					responseUser.OwnerID = ownerid

					responseUserA = append(responseUserA, responseUser)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseUserJSON, err := json.Marshal(responseUserA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseUserJSON)
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

func getRegistrationSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseRegistration ResponsRegistrationT
			var responseRegistrationA ResponseRegistrationAT

			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT id, type, status, description, user_id FROM registration`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
				var registrationList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						registrationList = append(registrationList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range registrationList {
					args = append(args, item)
				}
				if len(registrationList) > 0 {
					scanner = Session.Query(`SELECT id, type, status, description, user_id FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseRegistrationJSON, err := json.Marshal(responseRegistrationA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseRegistrationJSON)
						return
					}
				}
			} else {
				var registrationList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						registrationList = append(registrationList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range registrationList {
					args = append(args, item)
				}
				if len(registrationList) > 0 {
					scanner = Session.Query(`SELECT id, type, status, description, user_id FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseRegistrationJSON, err := json.Marshal(responseRegistrationA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseRegistrationJSON)
						return
					}
				}
			}
			for scanner.Next() {
				var id string
				var description string
				var typeMod string
				var status int
				var userID string

				err := scanner.Scan(&id, &typeMod, &status, &description, &userID)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseRegistration.ID = id
					responseRegistration.TypeMod = typeMod
					responseRegistration.Status = status
					responseRegistration.Description = description
					responseRegistration.UserID = userID

					responseRegistrationA = append(responseRegistrationA, responseRegistration)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseRegistrationJSON, err := json.Marshal(responseRegistrationA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseRegistrationJSON)
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

func putRegistrationUpdate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the module update!", claims["username"])

	type Registration struct {
		ID          string `json:"id"`
		TypeMod     string `json:"type"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			log.Print(string(b))
			if err := json.Unmarshal(b, &registration); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE registration SET type = ?, status = ?, description = ? WHERE id = ?`, registration.TypeMod, registration.Status, registration.Description, registration.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to change the module data (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully changed the module data %v!", claims["username"], registration.ID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putRegistrationCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of the module!", claims["username"])

	type Registration struct {
		ID          string `json:"id"`
		TypeMod     string `json:"type"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			id, err := gocql.RandomUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			if err := json.Unmarshal(b, &registration); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
				BEGIN BATCH
					INSERT INTO registration (id, type, status, description, user_id) VALUES (?, ?, ?, ?, ?)
					INSERT INTO user_registration (user_id, registration_id) VALUES (?, ?)
				APPLY BATCH
				`, id, registration.TypeMod, registration.Status, registration.Description, claims["userid"], claims["userid"], id).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a module (BATCH: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully created a module!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to create a module!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putRegistrationRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the removal of the module!", claims["username"])

	type Registration struct {
		ID string `json:"id"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			if err := json.Unmarshal(b, &registration); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()

				var userID string
				if err := Session.Query(`SELECT user_id FROM registration WHERE id = ? LIMIT 1`, registration.ID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&userID); err != nil {
					log.Print("Failed to remove the module (SELECT user_id: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				err := Session.Query(`
				BEGIN BATCH
					DELETE FROM registration WHERE id = ?
					DELETE FROM user_registration WHERE user_id = ? AND registration_id = ?
				APPLY BATCH
			`, registration.ID, userID, registration.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to remove the module (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully deleted the module %v!", claims["username"], registration.ID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to remove the module!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func getTaskSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var responseTask ResponseTaskT
			var responseTaskA ResponseTaskAT

			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type FROM task`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseTaskJSON, err := json.Marshal(responseTaskA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseTaskJSON)
						return
					}
				}
			} else {
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseTaskJSON, err := json.Marshal(responseTaskA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseTaskJSON)
						return
					}
				}
			}

			for scanner.Next() {
				var moduleID string
				var spaceID string
				var object string
				var metric string
				var status int
				var critical string
				var warning string
				var interval int64
				var dataType string

				err := scanner.Scan(&moduleID, &spaceID, &object, &metric, &status, &critical, &warning, &interval, &dataType)
				if err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseTask.ModuleID = moduleID
					responseTask.SpaceID = spaceID
					responseTask.Object = object
					responseTask.Metric = metric
					responseTask.Status = status
					responseTask.Critical = critical
					responseTask.Warning = warning
					responseTask.Interval = interval
					responseTask.DataType = dataType

					scannerModule := Session.Query(`SELECT description FROM registration WHERE id = ?`, moduleID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var moduleDesc string
					for scannerModule.Next() {
						err := scannerModule.Scan(&moduleDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseTask.ModuleDesc = moduleDesc
						}
					}

					scannerSpace := Session.Query(`SELECT description FROM space WHERE id = ?`, spaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var spaceDesc string
					for scannerSpace.Next() {
						err := scannerSpace.Scan(&spaceDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseTask.SpaceDesc = spaceDesc
						}
					}

					responseTaskA = append(responseTaskA, responseTask)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseTaskJSON, err := json.Marshal(responseTaskA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseTaskJSON)
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

func putTaskUpdate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the update of the problem!", claims["username"])

	type Task struct {
		ModuleID string `json:"moduleid"`
		SpaceID  string `json:"spaceid"`
		Object   string `json:"object"`
		Metric   string `json:"metric"`
		Status   int    `json:"status"`
		Critical string `json:"critical"`
		Warning  string `json:"warning"`
		Interval int64  `json:"interval"`
		DataType string `json:"datatype"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE task SET status = ?, critical = ?, warning = ?, interval = ?, data_type = ? WHERE module_id = ? AND space_id = ? AND object = ? AND metric = ?`, task.Status, task.Critical, task.Warning, task.Interval, task.DataType, task.ModuleID, task.SpaceID, task.Object, task.Metric).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to change these tasks (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully changed these tasks for the module %v!", claims["username"], task.ModuleID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putTaskCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of the task!", claims["username"])

	type Task struct {
		ModuleID string `json:"moduleid"`
		SpaceID  string `json:"spaceid"`
		Object   string `json:"object"`
		Metric   string `json:"metric"`
		Status   int    `json:"status"`
		Critical string `json:"critical"`
		Warning  string `json:"warning"`
		Interval int64  `json:"interval"`
		DataType string `json:"datatype"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`INSERT INTO task (module_id, space_id, object, metric, status, int_id, critical, warning, interval, data_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, task.ModuleID, task.SpaceID, task.Object, task.Metric, task.Status, rand.Int(), task.Critical, task.Warning, task.Interval, task.DataType).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a task (INSERT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully created the task!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to create a task!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putTaskRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the removal of the task!", claims["username"])

	type Task struct {
		ModuleID string `json:"moduleid"`
		SpaceID  string `json:"spaceid"`
		Object   string `json:"object"`
		Metric   string `json:"metric"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`DELETE FROM task WHERE module_id = ? AND space_id = ? AND object = ? AND metric = ?`, task.ModuleID, task.SpaceID, task.Object, task.Metric).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to delete the task (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully deleted the module task %v!", claims["username"], task.ModuleID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to delete the task!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func getSpaceSelect(w http.ResponseWriter, r *http.Request) {
	type ResponseT struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		UserID      string `json:"userid"`
	}
	type ResponseAT []ResponseT

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var response ResponseT
			var responseA ResponseAT

			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT id, description, status, user_id FROM space`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
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
				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT id, description, status, user_id FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseJSON, err := json.Marshal(responseA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseJSON)
						return
					}
				}
			} else {
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
				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT id, description, status, user_id FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseJSON, err := json.Marshal(responseA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseJSON)
						return
					}
				}
			}
			for scanner.Next() {
				var id string
				var description string
				var status int
				var userID string

				err := scanner.Scan(&id, &description, &status, &userID)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					response.ID = id
					response.Description = description
					response.Status = status
					response.UserID = userID

					responseA = append(responseA, response)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseJSON, err := json.Marshal(responseA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
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

func putSpaceCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of space!", claims["username"])

	type Space struct {
		Description string `json:"description"`
		Status      int    `json:"status"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			id, err := gocql.RandomUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var space Space
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
				BEGIN BATCH
					INSERT INTO space (id, description, status, user_id) VALUES (?, ?, ?, ?)
					INSERT INTO user_space (user_id, space_id) VALUES (?, ?)
				APPLY BATCH
				`, id, space.Description, space.Status, claims["userid"], claims["userid"], id).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a space (INSERT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				log.Printf("User %v successfully created the space!", claims["username"])
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			log.Print("It was not possible to create a space!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putSpaceUpdate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the update of space!", claims["username"])

	type Space struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Status      int    `json:"status"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var space Space
			log.Print(string(b))
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE space SET description = ?, status = ? WHERE id = ?`, space.Description, space.Status, space.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to change these space (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully changed these spaces %v!", claims["username"], space.ID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putSpaceRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the removal of space!", claims["username"])

	type Space struct {
		ID string `json:"id"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var space Space
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				var userID string
				if err := Session.Query(`SELECT user_id FROM space WHERE id = ? LIMIT 1`, space.ID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&userID); err != nil {
					log.Print("Failed to remove space (SELECT user_id: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
				}

				err := Session.Query(`
				BEGIN BATCH 
					DELETE FROM space WHERE id = ?
					DELETE FROM user_space WHERE user_id = ? AND space_id = ?
				APPLY BATCH
			`, space.ID, userID, space.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to remove space (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				log.Printf("User %v successfully removed the space %v!", claims["username"], space.ID)
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			log.Print("It was not possible to remove the space!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func getRawDataCount(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM raw_data01`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
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
			} else {
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

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM raw_text01`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
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
			} else {
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

func putChartData(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested scheduled data", claims["username"])

	type RequestT struct {
		DataSRC string `json:"datasrc"`
	}

	type ChartDataSetT struct {
		BackgroundColor string    `json:"backgroundColor"`
		Data            []float32 `json:"data"`
		BarPercentage   float32   `json:"barPercentage"` // Пространство между столбиками
	}

	type ChartDataT struct {
		Labels   []int64         `json:"labels"`
		DataSets []ChartDataSetT `json:"datasets"`
	}

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

			var chartDataSet ChartDataSetT
			var chartData ChartDataT
			ctx := context.Background()
			var scanner gocql.Scanner
			stopTime := chrdsclient.MakeTimestamp()
			startTime := stopTime - (23 * 60 * 60 * 1000)

			chartDataSet.BackgroundColor = "#ade2ffbe"
			chartDataSet.BarPercentage = 1.1 // Пространство между столбиками

			var spaceID []string

			{
				if claims["role"] == "superadmin" {
					scanner = Session.Query(`SELECT id, status FROM space`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else if claims["role"] == "admin" {
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
					if len(spaceList) > 0 {
						scanner = Session.Query(`SELECT id, status FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					}
				} else {
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
					if len(spaceList) > 0 {
						scanner = Session.Query(`SELECT id, status FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					}
				}

				for scanner.Next() {
					var id string
					var status int
					err := scanner.Scan(&id, &status)
					if err != nil {
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if status == 1 {
						spaceID = append(spaceID, id)
					}
				}
			}

			if len(spaceID) > 0 {
				for iTime := startTime; iTime <= stopTime; iTime = iTime + (60 * 60 * 1000) {
					chartData.Labels = append(chartData.Labels, iTime)

					args := []interface{}{}
					for _, item := range spaceID {
						args = append(args, item)
					}
					args = append(args, iTime)
					args = append(args, iTime+(60*60*1000))

					syntKeyList := makeDateList(1)
					for _, item := range syntKeyList {
						args = append(args, item)
					}

					var scanner gocql.Scanner
					if request.DataSRC == "raw_text" {
						scanner = Session.Query(`SELECT count(*) FROM raw_text01 WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceID)-1)+`) AND event_time >= ? AND event_time < ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					} else if request.DataSRC == "raw_data" {
						scanner = Session.Query(`SELECT count(*) FROM raw_data01 WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceID)-1)+`) AND event_time >= ? AND event_time < ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					} else {
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					for scanner.Next() {
						var count int64
						err := scanner.Scan(&count)
						if err != nil {
							go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
							w.WriteHeader(http.StatusInternalServerError)
							return
						} else {
							chartDataSet.Data = append(chartDataSet.Data, float32(count))
						}
					}
					if err := scanner.Err(); err != nil {
						log.Print(err)
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}

			chartData.DataSets = append(chartData.DataSets, chartDataSet)

			responseJSON, err := json.Marshal(chartData)
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
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Print("Failed to get data!")
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
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

func getDatabaseStatus(w http.ResponseWriter, r *http.Request) {
	type ResponseT struct {
		Peer       string  `json:"peer"`
		Datacenter string  `json:"datacenter"`
		HostID     string  `json:"hostid"`
		Owns       float32 `json:"owns"`
		Tokens     int     `json:"tokens"`
		Load       string  `json:"load"`
		Status     string  `json:"status"`
		Up         bool    `json:"up"`
	}
	type ResponseAT []ResponseT

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var response ResponseT
			var responseA ResponseAT

			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT peer, dc, host_id, owns, tokens, load, status, up FROM system.cluster_status`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			}

			for scanner.Next() {
				var peer string
				var dc string
				var hostID string
				var owns float32
				var tokens int
				var load string
				var status string
				var up bool

				err := scanner.Scan(&peer, &dc, &hostID, &owns, &tokens, &load, &status, &up)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					response.Peer = peer
					response.Datacenter = dc
					response.HostID = hostID
					response.Owns = owns
					response.Tokens = tokens
					response.Load = load
					response.Status = status
					response.Up = up

					responseA = append(responseA, response)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseJSON, err := json.Marshal(responseA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
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

func getDataManagerStatus(w http.ResponseWriter, r *http.Request) {
	type ResponseT struct {
		Up bool `json:"up"`
	}
	type ResponseAT []ResponseT

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var response ResponseT
			var responseA ResponseAT

			status := chrdsclient.Status()

			for _, item := range status {
				response.Up = item
				responseA = append(responseA, response)
			}

			responseJSON, err := json.Marshal(responseA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
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

func getDashboardCount(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM dashboard WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else {
				scanner = Session.Query(`SELECT count(*) FROM dashboard WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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

func getProblemCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM problem`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM problem WHERE module_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
			} else {
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM problem WHERE module_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
					log.Print(err)
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

func getProblemSelect(w http.ResponseWriter, r *http.Request) {
	type ResponseProblemT struct {
		ModuleID       string `json:"moduleid"`
		SpaceID        string `json:"spaceid"`
		Metric         string `json:"metric"`
		Status         string `json:"status"`
		EventTime      int64  `json:"eventtime"`
		EventTimeStart int64  `json:"eventtimestart"`
		Value          string `json:"value"`
		ModuleDesc     string `json:"moduledesc"`
		SpaceDesc      string `json:"spacedesc"`
	}
	type ResponseProblemAT []ResponseProblemT

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var responseProblem ResponseProblemT
			var responseProblemA ResponseProblemAT

			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT space_id, module_id, metric, event_time, event_time_start, status, value FROM problem`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else if claims["role"] == "admin" {
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT space_id, module_id, metric, event_time, event_time_start, status, value FROM problem WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseProblemJSON, err := json.Marshal(responseProblemA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseProblemJSON)
						return
					}
				}
			} else {
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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

				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT space_id, module_id, metric, event_time, event_time_start, status, value FROM problem WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Iter().Scanner()
				} else {
					responseProblemJSON, err := json.Marshal(responseProblemA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseProblemJSON)
						return
					}
				}
			}

			for scanner.Next() {
				var moduleID string
				var spaceID string
				var metric string
				var status string
				var eventTime int64
				var eventTimeStart int64
				var value string

				err := scanner.Scan(&spaceID, &moduleID, &metric, &eventTime, &eventTimeStart, &status, &value)
				if err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseProblem.ModuleID = moduleID
					responseProblem.SpaceID = spaceID
					responseProblem.Metric = metric
					responseProblem.Status = status
					responseProblem.EventTime = eventTime
					responseProblem.EventTimeStart = eventTimeStart
					responseProblem.Value = value

					scannerModule := Session.Query(`SELECT description FROM registration WHERE id = ?`, moduleID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var moduleDesc string
					for scannerModule.Next() {
						err := scannerModule.Scan(&moduleDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseProblem.ModuleDesc = moduleDesc
						}
					}

					scannerSpace := Session.Query(`SELECT description FROM space WHERE id = ?`, spaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var spaceDesc string
					for scannerSpace.Next() {
						err := scannerSpace.Scan(&spaceDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseProblem.SpaceDesc = spaceDesc
						}
					}

					responseProblemA = append(responseProblemA, responseProblem)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseProblemJSON, err := json.Marshal(responseProblemA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseProblemJSON)
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

func getDashboardSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a list of dashboard", claims["username"])

	type ResponseChartDataT struct {
		ID         string `json:"id"`
		Metric     string `json:"metric"`
		SpaceID    string `json:"spaceid"`
		GraphType  string `json:"graphtype"`
		GraphColor string `json:"graphcolor"`
		GroupFunc  string `json:"groupfunc"` // Max, Min, Avg, Sum
		GraphOrder int    `json:"graphorder"`
	}

	type ResponseDashboardT struct {
		ID        string               `json:"id"`
		Name      string               `json:"name"`
		StartTime int64                `json:"starttime"`
		StopTime  int64                `json:"stoptime"`
		Status    int                  `json:"status"`
		ChartData []ResponseChartDataT `json:"chartdata"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		var responseChartData ResponseChartDataT
		var responseDashboard ResponseDashboardT
		var responseDashboardA []ResponseDashboardT

		ctx01 := context.Background()
		scanner01 := Session.Query(`SELECT id, name, start_time, stop_time, status FROM dashboard WHERE user_id = ?`, claims["userid"]).WithContext(ctx01).Consistency(ConsistencyRead).Iter().Scanner()
		for scanner01.Next() {
			var dashboardID string
			var dashboardName string
			var startTime int64
			var stopTime int64
			var dashboardStatus int
			err := scanner01.Scan(&dashboardID, &dashboardName, &startTime, &stopTime, &dashboardStatus)
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				if dashboardStatus == 1 {
					var responseChartDataA []ResponseChartDataT
					ctx02 := context.Background()
					scanner02 := Session.Query(`SELECT id, graph_color, graph_func, graph_type, metric, space_id, status, graph_order FROM dashboard_chartdata WHERE dashboard_id = ?`, dashboardID).WithContext(ctx02).Consistency(ConsistencyRead).Iter().Scanner()
					for scanner02.Next() {
						var chartDataID string
						var graphColor string
						var graphFunc string
						var graphType string
						var metric string
						var spaceID string
						var chartDataStatus int
						var graphOrder int
						err := scanner02.Scan(&chartDataID, &graphColor, &graphFunc, &graphType, &metric, &spaceID, &chartDataStatus, &graphOrder)
						if err != nil {
							log.Print(err)
							w.WriteHeader(http.StatusInternalServerError)
							return
						} else {
							responseChartData.ID = chartDataID
							responseChartData.Metric = metric
							responseChartData.SpaceID = spaceID
							responseChartData.GraphColor = graphColor
							responseChartData.GraphType = graphType
							responseChartData.GroupFunc = graphFunc
							responseChartData.GraphOrder = graphOrder

							responseChartDataA = append(responseChartDataA, responseChartData)
						}
					}
					responseDashboard.ID = dashboardID
					responseDashboard.Name = dashboardName
					responseDashboard.StartTime = startTime
					responseDashboard.StopTime = stopTime
					responseDashboard.Status = dashboardStatus
					responseDashboard.ChartData = responseChartDataA

					responseDashboardA = append(responseDashboardA, responseDashboard)
				}
			}
		}

		responseJSON, err := json.Marshal(responseDashboardA)
		if err != nil {
			log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(responseJSON)
			return
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func putChartDashboardData(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested Dashboard data", claims["username"])

	type RequesChartDataT struct {
		ID         string `json:"id"`
		Metric     string `json:"metric"`
		SpaceID    string `json:"spaceid"`
		GraphType  string `json:"graphtype"`
		GraphColor string `json:"graphcolor"`
		GroupFunc  string `json:"groupfunc"` // Max, Min, Avg, Sum
		GraphOrder int    `json:"graphorder"`
	}

	type RequestChartDataSetT struct {
		ID        string             `json:"id"`
		Name      string             `json:"name"`
		StartTime int64              `json:"starttime"`
		StopTime  int64              `json:"stoptime"`
		ChartData []RequesChartDataT `json:"chartdata"`
	}

	type ChartDataSetT struct {
		Label           string    `json:"label"`
		Type            string    `json:"type"`
		BackgroundColor string    `json:"backgroundColor"`
		BorderColor     string    `json:"borderColor"`
		Data            []float32 `json:"data"`
		BarPercentage   float32   `json:"barPercentage"` // The space between the bars
		Order           int       `json:"order"`
	}

	type ChartDataT struct {
		Labels    []int64         `json:"labels"`
		DataSets  []ChartDataSetT `json:"datasets"`
		StartTime int64           `json:"starttime"`
		StopTime  int64           `json:"stoptime"`
	}

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

			var request []RequestChartDataSetT
			if err := json.Unmarshal(b, &request); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var chartDataA []ChartDataT

			for _, requestA := range request {
				var chartDataSet ChartDataSetT
				var chartData ChartDataT

				var stopTime int64 = 0   // The end of the sample period
				var startTime int64 = 0  // The start time of the sample period
				var deltaTime int64 = 60 // Group step by sampling time in seconds
				ctx := context.Background()
				if requestA.StopTime == 0 {
					stopTime = chrdsclient.MakeTimestamp() - (60 * 1000)
				} else {
					stopTime = requestA.StopTime
				}
				if requestA.StartTime == 0 {
					startTime = stopTime - (30 * 60 * 1000) // Default sample period 30 minutes
				} else {
					startTime = requestA.StartTime
				}

				log.Print(startTime, " : ", stopTime)

				if stopTime-startTime <= (3 * 60 * 1000) { // Up to 3 minutes
					deltaTime = 1 // 1 second
				} else if stopTime-startTime > (3*60*1000) && stopTime-startTime <= (60*60*1000) { // up to 1 hour
					deltaTime = 60 // 1 minute
				} else if stopTime-startTime > (60*60*1000) && stopTime-startTime <= (120*60*1000) { // Up to 2 hours
					deltaTime = 600 // 10 minutes
				} else if stopTime-startTime > (120*60*1000) && stopTime-startTime <= (1440*60*1000) { // Up to 24 hours
					deltaTime = 900 // 15 minutes
				} else if stopTime-startTime > (1440*60*1000) && stopTime-startTime <= (4320*60*1000) { // Up to 3 days
					deltaTime = 3600 // 1 time
				} else { // More than 3 days
					deltaTime = 86400 // 1 day
				}

				chartData.StartTime = startTime
				chartData.StopTime = stopTime

				for iTime := startTime; iTime <= stopTime; iTime = iTime + (deltaTime * 1000) {
					chartData.Labels = append(chartData.Labels, iTime)
				}

				for _, reqChartData := range requestA.ChartData {
					chartDataSet.Label = reqChartData.Metric
					chartDataSet.Type = reqChartData.GraphType
					if reqChartData.GraphColor == "" {
						chartDataSet.BackgroundColor = "#ade2ffbe" // Default color: #ade2ffbe  ade2ffbe
					} else {
						chartDataSet.BackgroundColor = reqChartData.GraphColor
					}
					chartDataSet.BorderColor = reqChartData.GraphColor
					chartDataSet.BarPercentage = 0.9 // The space between the bars
					chartDataSet.Order = reqChartData.GraphOrder
					chartDataSet.Data = nil

					for _, iTime := range chartData.Labels {
						args := []interface{}{}
						args = append(args, reqChartData.SpaceID)
						args = append(args, reqChartData.Metric)
						args = append(args, iTime)
						args = append(args, iTime+(deltaTime*1000))

						syntKeyList := makeDateList(int(((stopTime - startTime) / 1000 / 60 / 60 / 24 / 30) + 1)) // Preparation of the list of keys for data sample
						for _, item := range syntKeyList {
							args = append(args, item)
						}

						scanner := Session.Query(`SELECT value FROM raw_data02 WHERE space_id = ? AND metric = ? AND event_time >= ? AND event_time < ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()

						var valueCount float32 = 0
						var valueSum float32 = 0
						for scanner.Next() {
							var value float32
							err := scanner.Scan(&value)
							if err != nil {
								go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
								w.WriteHeader(http.StatusInternalServerError)
								return
							} else {
								valueSum = valueSum + value
							}
							valueCount++
						}
						if err := scanner.Err(); err != nil {
							log.Print(err)
							go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
							w.WriteHeader(http.StatusInternalServerError)
							return
						}

						if valueCount == 0 {
							chartDataSet.Data = append(chartDataSet.Data, 0)
						} else {
							chartDataSet.Data = append(chartDataSet.Data, valueSum/valueCount)
						}
					}

					chartData.DataSets = append(chartData.DataSets, chartDataSet)
				}

				chartDataA = append(chartDataA, chartData)
			}

			responseJSON, err := json.Marshal(chartDataA)
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
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Print("Failed to get data!")
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putDashboardCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of a dashboard!", claims["username"])

	type Dashboard struct {
		Name      string `json:"name"`
		StartTime int64  `json:"starttime"`
		StopTime  int64  `json:"stoptime"`
		Status    int    `json:"status"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := gocql.RandomUUID()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var dashboard Dashboard
		if err := json.Unmarshal(b, &dashboard); err != nil {
			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		{
			ctx := context.Background()
			err := Session.Query(`
			INSERT INTO dashboard (user_id, id, name, start_time, stop_time, status) VALUES (?, ?, ?, ?, ?, ?)
			`, claims["userid"], id, dashboard.Name, dashboard.StartTime, dashboard.StopTime, dashboard.Status).WithContext(ctx).Exec()
			if err != nil {
				log.Print("Failed to create a dashboard (ERROR: ", err.Error(), ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				log.Printf("User %v successfully created Dashboard!", claims["username"])
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	} else {
		log.Print("It was not possible to create a dashboard!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func putDashboardEdit(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested dashboard update!", claims["username"])

	type DashboardChartData struct {
		ID         string `json:"id"`
		Metric     string `json:"metric"`
		SpaceID    string `json:"spaceid"`
		GraphType  string `json:"graphtype"`
		GraphColor string `json:"graphcolor"`
		GroupFunc  string `json:"groupfunc"` // Max, Min, Avg, Sum
		GraphOrder int    `json:"graphorder"`
	}

	type Dashboard struct {
		ID        string               `json:"id"`
		Name      string               `json:"name"`
		StartTime int64                `json:"starttime"`
		StopTime  int64                `json:"stoptime"`
		Status    int                  `json:"status"`
		ChartData []DashboardChartData `json:"chartdata"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.Background()

		var dashboard Dashboard
		if err := json.Unmarshal(b, &dashboard); err != nil {
			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = Session.Query(`
		BEGIN BATCH
			DELETE FROM dashboard_chartdata WHERE dashboard_id = ?
			DELETE FROM dashboard WHERE user_id = ? AND id = ?
		APPLY BATCH
		`, dashboard.ID, claims["userid"], dashboard.ID).WithContext(ctx).Exec()
		if err != nil {
			log.Print("Failed to remove dashboard (ERROR: ", err.Error(), ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			uuid, err := gocql.RandomUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = Session.Query(`
			INSERT INTO dashboard (user_id, id, name, start_time, status, stop_time) VALUES (?, ?, ?, ?, ?, ?)
			`, claims["userid"], uuid, dashboard.Name, dashboard.StartTime, 1, dashboard.StopTime).WithContext(ctx).Exec()
			if err != nil {
				log.Print("Failed to update dashboard (ERROR: ", err.Error(), ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				for _, v := range dashboard.ChartData {
					uuidChartData, err := gocql.RandomUUID()
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					err = Session.Query(`
					INSERT INTO dashboard_chartdata (dashboard_id, graph_order, id, graph_color, graph_func, graph_type, metric, space_id, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
					`, uuid, v.GraphOrder, uuidChartData, v.GraphColor, v.GroupFunc, v.GraphType, v.Metric, v.SpaceID, 1).WithContext(ctx).Exec()
					if err != nil {
						log.Print("Failed to update dashboarding graphs (ERROR: ", err.Error(), ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				log.Printf("User %v successfully created dashboard!", claims["username"])
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}
}

func putDashboardRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested dashboard removal!", claims["username"])

	type Dashboard struct {
		ID string `json:"id"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var dashboard Dashboard
		if err := json.Unmarshal(b, &dashboard); err != nil {
			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		{
			ctx := context.Background()
			err = Session.Query(`
			BEGIN BATCH
				DELETE FROM dashboard_chartdata WHERE dashboard_id = ?
				DELETE FROM dashboard WHERE user_id = ? AND id = ?
			APPLY BATCH
			`, dashboard.ID, claims["userid"], dashboard.ID).WithContext(ctx).Exec()
			if err != nil {
				log.Print("Failed to remove dashboard (ERROR: ", err.Error(), ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			log.Printf("User %v successfully deleted Dashboard!", claims["username"])
			w.WriteHeader(http.StatusOK)
			return
		}
	} else {
		log.Print("Failed to remove dashboard (API VERSION ERROR)!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
