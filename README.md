# Charybdis Monitoring System UIManager 1.0.5

The **Charybdis Monitoring System** project is an attempt to create a simple infrastructure and application monitoring system based on Zabbix's best practices while addressing its main weaknesses in terms of scalability and data storage.  
**UIManager** is a component that implements a full-fledged user WEB interface for the monitoring system's functions.

<img width="960" alt="Ch01" src="https://github.com/user-attachments/assets/0b1e77b6-67d8-4c74-8925-613ca2c176c5" />

## Installing UIManager
It is recommended to install **UIManager** after installing **DataManager**.

### Preparing the Environment
Download the latest release or clone the source code from the **master** branch:
```sh
git clone https://github.com/DemianSV/chrdsuimanager.git
```

### Installing UIManager from a Docker Image
**Building and Running**  
Navigate to the directory with the cloned repository.

Build the image (edit the *Dockerfile* if necessary):
```sh
docker build -t chrdsuimanager
```

Configure the environment variables and launch parameters in the *docker-compose.yml* file.

Start **UIManager** by running:
```sh
docker compose up -d
```

Verify the result:
```sh
docker ps -a
docker logs chrdsuimanager01
```
Try opening **UIManager** in a browser. For a local installation, the address will be: http://localhost:7007. If installed on a remote server, adjust the URL accordingly.

**Stopping**  
```sh
docker compose down
```

### Installing UIManager Directly on a Server
**Building and Running**  
Install the latest version of GoLang (if not already installed) from the official website (https://go.dev).  
Install the latest version of NodeJS from the official website (https://nodejs.org/en/download).

Navigate to the directory with the cloned repository.

Build **UIManager**:
```sh
cd frontend
npm install -D vite
npm install
npm run build
cp -r dist/assets ../
cp dist/index.html ../
cd ../
go build
```

Configure the parameters in the *chrdsuimanager.json* file.

Run **UIManager**:
```sh
./chrdsuimanager
# or
./chrdsuimanager -log file
```

### Configuration Parameters
Configuration can be done in two ways:
1. Via the configuration file (chrdsuimanager.json)
2. Via environment variables (for the Docker option)

>**Environment variables have higher priority and will always override parameters from the configuration file!**

>**Variable naming convention:**  
>**CHRDS**: Indicates the variable belongs to the monitoring system,  
>**HTTP**: Group of variables for the HTTP scope,  
>**HOST**: Configuration parameter.

**CHRDS_UIMANAGER_MODULEID**: Reserved module ID for UIManager,  
**CHRDS_UIMANAGER_SPACEID**: Reserved space ID for UIManager,  
**CHRDS_UIMANAGER_DATAMANAGERURL**: Array of DataManager API URLs.  
**CHRDS_UIMANAGER_DATAMANAGERTIMEOUT**: Timeout for DataManager API requests.  
**CHRDS_UIMANAGER_AUDITCHECK**: Mandatory audit availability check during authorization (true).  

**CHRDS_HTTP_HOST**: Address for incoming HTTP server connections (API, Swagger),  
**CHRDS_HTTP_PORT**: Port for incoming HTTP server connections,  
**CHRDS_HTTP_READTIMEOUT**: HTTP server read timeout,  
**CHRDS_HTTP_WRITETIMEOUT**: HTTP server write timeout,  
**CHRDS_HTTP_RL**: Requests Limit, HTTP server incoming request limit per second,  
**CHRDS_HTTP_TLS**: Enable TLS for HTTP server (true/false),  
**CHRDS_HTTP_CERTPATH**: Path to the certificate for HTTP server connections,  
**CHRDS_HTTP_KEYPATH**: Path to the private key for HTTP server connections,  
**CHRDS_HTTP_CAPATH**: Path to the CA certificate for HTTP server connections,  
**CHRDS_HTTP_CLIENTINSECURE**: Disable server certificate verification for outgoing connections (true/false).  

**CHRDS_DB_HOST**: Array of database node addresses, comma-separated (e.g., 127.0.0.1:9042,127.0.0.1:9043),  
**CHRDS_DB_USERNAME**: Database username,  
**CHRDS_DB_PASSWORD**: Database password,  
**CHRDS_DB_TLS**: Enable TLS for database cluster interactions (true/false),  
**CHRDS_DB_CERTPATH**: Path to the certificate for database connections,  
**CHRDS_DB_KEYPATH**: Path to the private key for database connections,  
**CHRDS_DB_CAPATH**: Path to the CA certificate for database connections,  
**CHRDS_DB_HOSTVERIFICATION**: Enable HostVerification for database connections (true/false),  
**CHRDS_DB_KEYSPACE**: KeySpace for the connection,  
**CHRDS_DB_CONSISTENCY**: Consistency parameter for database data,  
**CHRDS_DB_CONSISTENCYREAD**: Consistency parameter for database queries,  
**CHRDS_DB_TIMEOUT**: Query execution timeout before forced termination.  

**Configuration via chrdsuimanager.json File**  
>Parameter groups and parameters in the file are identical to those in the environment variables.  

>When using the configuration file, remember that its parameters have lower priority than environment variables and will be overridden by them!  

>Combining parameters from different sources is allowed, keeping the above in mind.  

```json
{
	"UIMANAGER": {
		"MODULEID": "623a57c2-3df5-4287-ada5-82f7e4a0b5db",
		"SPACEID": "17f0bd20-41cf-4801-a481-ff721a41fa93",
		"DATAMANAGERURL": ["http://172.20.0.10:6006", "http://172.20.0.11:6006"]
	},
	"HTTP": {
		"HOST": "0.0.0.0",
		"PORT": "7007",
		"READTIMEOUT": 30,
		"WRITETIMEOUT": 30,
		"RL": 100,
		"TLS": false,
		"CERTPATH": "",
		"KEYPATH": "",
		"CAPATH": "",
		"CLIENTINSECURE": true
	},
	"DB": {
		"HOSTS": ["127.0.0.1:9042", "127.0.0.1:9043"],
		"USERNAME": "chrds",
		"PASSWORD": "",
		"TLS": false,
		"CERTPATH": "",
		"KEYPATH": "",
		"CAPATH": "",
		"HOSTVERIFICATION": false,
		"KEYSPACE": "chrds",
		"CONSISTENCY": "Quorum",
		"TIMEOUT": 30000
	}
}
```

## Using Grafana
Create a user in the Cassandra database, for example **Grafana**, and give out the right only to read for the table **chrds.raw_data02**.

Install Grafana in any convenient way using the documentation from the official website (https://grafana.com/docs/grafana/latest/setup-grafana/installation/).

Turn on and configure Data Sources Plugin: **Apache Cassandra Datasource for Grafana**.

<img width="1453" height="243" alt="Снимок экрана 2025-08-30 в 18 25 56" src="https://github.com/user-attachments/assets/ac9cc4f9-ea1e-4f4b-a314-7f161624c361" />

Set up a plugin to connect to a database with an previously established UZ.

Create a dashboard on the metrics stored in the database using an example of a request:

```sql
SELECT space_id, value, totimestamp(maxtimeuuid(event_time)), space_description FROM chrds.raw_data02 WHERE space_id IN (65c6b051-10fb-4bd7-8c04-7fe478e55d13, 34bcd935-8e7a-4b79-b76b-352cf3ece91f) AND metric = 'system.cpu.util' AND synt_key IN ('${__from:date:YYYY.MM}', '${__to:date:YYYY.MM}') AND event_time > $__from and event_time < $__to
```

<img width="1134" height="874" alt="Снимок экрана 2025-09-13 в 10 31 56" src="https://github.com/user-attachments/assets/a2143064-11a2-4fc7-aa2e-f34a35e063c5" />

<img width="1469" height="820" alt="Снимок экрана 2025-09-13 в 10 26 15" src="https://github.com/user-attachments/assets/b2ac7f9f-b7b8-454f-af06-ed785bb50f3b" />


