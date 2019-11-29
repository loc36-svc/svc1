package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/go-sql-driver/mysql"
	sdsLib "github.com/loc36-core/sdsLib"
	"github.com/loc36-svc/svc1-http--cntlr"
	"github.com/loc36-svc/svc1-http--lib"
	"github.com/nicholoid-dtp/logBook"
	"github.com/qamarian-dtp/err"
	errLib "github.com/qamarian-lib/err"
	"gopkg.in/qamarian-lib/str.v3"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)
func init () {
	errU := sdsLib.InitReport ()
	if errU != nil {
		errX := err.New (`Package "github.com/loc36-core/sdsLib" init failed.`, nil, nil, sdsLib.InitReport ())
		str.PrintEtr (errLib.Fup (errX), "err", "main ()")
		os.Exit (1)
	}

	errV := cntlr.InitReport ()
	if errV != nil {
		errX := err.New (`Package "github.com/loc36-svc/svc1-http--cntlr" init failed.`, nil, nil, cntlr.InitReport ())
		str.PrintEtr (errLib.Fup (errX), "err", "main ()")
		os.Exit (1)
	}
}

// HTTP interface.
func main () {
	// Updating net addr with the SDS. ..1.. {
	port, errX := strconv.Atoi (httpConf ["net_port"])
	if errX != nil {
		errY := err.New ("Unable to convert HTTP port from string to int.", nil, nil, errX)
		str.PrintEtr (errLib.Fup (errY), "err", "main ()")
		os.Exit (1)
	}

	conn, errA := db.Conn (context.Background ())
	if errA != nil {
		errB := err.New ("Unable to source a conn to the SDS.", nil, nil, errA)
		str.PrintEtr (errLib.Fup (errB), "err", "main ()")
		os.Exit (1)
	}

	errC := sdsLib.UpdateAddr (httpConf ["net_addr"], port, serviceId, sds ["update_pass"], conn)
	if errC != nil {
		errD := err.New ("Unable to update addr with the SDS.", nil, nil, errC)
		str.PrintEtr (errLib.Fup (errD), "err", "main ()")
		os.Exit (1)
	}
	// ..1.. }

	// Creating the HTTP interface. ..1.. {
	readTimeout, errE       := time.ParseDuration (httpConf ["read_timeout"]        + "s")
	readHeaderTimeout, errF := time.ParseDuration (httpConf ["read_header_timeout"] + "s")
	writeTimeout, errG      := time.ParseDuration (httpConf ["wrte_timeout"]        + "s")
	idleTimeout, errH       := time.ParseDuration (httpConf ["idle_timeout"]        + "s")

	if errE != nil {
		errI := err.New ("Unable to create HTTP read timeout duration.", nil, nil, errE)
		str.PrintEtr (errLib.Fup (errI), "err", "main ()")
		os.Exit (1)
	}

	if errF != nil {
		errJ := err.New ("Unable to create HTTP read header timeout duration.", nil, nil, errF)
		str.PrintEtr (errLib.Fup (errJ), "err", "main ()")
		os.Exit (1)
	}

	if errG != nil {
		errK := err.New ("Unable to create HTTP write timeout duration.", nil, nil, errG)
		str.PrintEtr (errLib.Fup (errK), "err", "main ()")
		os.Exit (1)
	}

	if errH != nil {
		errL := err.New ("Unable to create HTTP idle timeout duration.", nil, nil, errH)
		str.PrintEtr (errLib.Fup (errL), "err", "main ()")
		os.Exit (1)
	}

	intf := &http.Server {
		Addr: fmt.Sprintf ("%s:%s", "0.0.0.0", httpConf ["net_port"]),
		ReadTimeout: readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout: idleTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	// ..1.. }

	// Create router. ..1.. {
	router := mux.NewRouter ()
	router.HandleFunc ("/report", cntlr.Report).Methods ("POST")
	// ..1.. }

	// Attaching router to the interface
	intf.Handler = router

	notf := fmt.Sprintf ("HTTP interface addr: %s:%s [HTTPS]", "0.0.0.0", httpConf ["net_port"])
	str.PrintEtr (notf, "std", "main ()")

	// Start HTTP interface.
	errQ := intf.ListenAndServeTLS (httpConf ["tls_crt"], httpConf ["tls_key"])

	if errQ != nil && errQ != http.ErrServerClosed {
		errR := err.New ("HTTP interface shutdown due to an error.", nil, nil, errQ)
		errMssg := fmt.Sprintf ("%s {main ()}", errLib.Fup (errR))
		logBk.Record ([]byte (errMssg))
		os.Exit (1)
	}
}
var (
	sds map[string]string       // SDS conf
	httpConf map[string]string  // HTTP interface conf
	db *sql.DB                  // Conn to SDS.
	serviceId = "1"
	logBk = logBook.New (os.Stderr) // Std err
)
func init () {
	// Loading conf. ..1.. {
	var errX error
	sds, httpConf, errX = lib.Conf ()
	if errX != nil {
		errY := err.New ("Unable to load conf.", nil, nil, errX)
		str.PrintEtr (errLib.Fup (errY), "err", "main ()")
		os.Exit (1)
	}
	// ..1.. }

	// Extracting pub key from file. ..1.. {
	fileContent, errY := ioutil.ReadFile (sds ["pub_key"])
	if errY != nil {
		errT := err.New ("Unable to read in sds pub key file.", nil, nil, errY)
		str.PrintEtr (errLib.Fup (errT), "err", "main ()")
		os.Exit (1)
	}

	block, _ := pem.Decode (fileContent)
	if block == nil || block.Type != "PUBLIC KEY" {
		errT := err.New ("SDS pub key file seems invalid.", nil, nil)
		str.PrintEtr (errLib.Fup (errT), "err", "main ()")
		os.Exit (1)
	}

	pubKey, errZ := x509.ParsePKIXPublicKey (block.Bytes)
	if errZ != nil {
		errT := err.New ("Unable to parse sds pub key.", nil, nil, errZ)
		str.PrintEtr (errLib.Fup (errT), "err", "main ()")
		os.Exit (1)
	}

	key, okA := pubKey.(*rsa.PublicKey)
	if okA == false {
		errT := err.New ("Result of sds pub key parsing is not a valid pub key.", nil, nil)
		str.PrintEtr (errLib.Fup (errT), "err", "main ()")
		os.Exit (1)
	}

	mysql.RegisterServerPubKey ("dbmsPubKey", key)
	// ..1.. }

	// Establishing a conn to the SDS. ..1.. {
	connURLFormat := "%s:%s@tcp(%s:%s)/service_addr?tls=skip-verify&serverPubKey=dbmsPubKey&timeout=480s&writeTimeout=480s&" +
		"readTimeout=480s"

	connURL := fmt.Sprintf (connURLFormat, url.QueryEscape (sds ["user_name"]), url.QueryEscape (sds ["user_pass"]),
		url.QueryEscape (sds ["net_addr"]), url.QueryEscape (sds ["net_port"]))
	
	var errB error
	db, errB = sql.Open ("mysql", connURL)
	if errB != nil {
		errT := err.New ("Unable to connect to the SDS.", nil, nil, errB)
		str.PrintEtr (errLib.Fup (errT), "err", "main ()")
		os.Exit (1)
	}
	errC := db.Ping ()
	if errC != nil {
		errT := err.New ("Unable to connect to the SDS.", nil, nil, errC)
		str.PrintEtr (errLib.Fup (errT), "err", "main ()")
		os.Exit (1)
	}
	// ..1.. }
}
