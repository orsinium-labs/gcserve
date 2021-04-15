package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type Server struct {
	Address  string
	Bucket   string
	Cred     string
	Username string
	Password string
	Debug    bool

	gcs    *storage.Client
	logger *zap.Logger
}

func (s *Server) Init() error {
	if s.Bucket == "" {
		return errors.New("--bucket required")
	}
	if s.Username == "" {
		return errors.New("--username required")
	}
	if s.Password == "" {
		return errors.New("--password required")
	}

	var err error
	ctx := context.Background()

	// init GCS client
	opts := make([]option.ClientOption, 0)
	if s.Cred != "" {
		opts = append(opts, option.WithCredentialsFile(s.Cred))
	}
	s.gcs, err = storage.NewClient(ctx, opts...)
	if err != nil {
		return err
	}

	// init logger
	cfg := zap.NewProductionConfig()
	if s.Debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	s.logger, err = cfg.Build()
	if err != nil {
		return err
	}
	return nil
}

func (s Server) handleError(w http.ResponseWriter, err error) {
	if err == storage.ErrObjectNotExist {
		s.logger.Debug("object not found")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s.logger.Error("error getting object", zap.Error(err))
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s Server) Handle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// authorize
	user, pass, ok := r.BasicAuth()
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	if !ok {
		http.Error(w, "Not authorized", 401)
		return
	}
	if user != s.Username || pass != s.Password {
		http.Error(w, "Not authorized", 401)
		return
	}

	// get object metainfo
	oname := strings.TrimPrefix(ps.ByName("object"), "/")
	if strings.HasSuffix(oname, "/") {
		oname += "index.html"
	}
	s.logger.Debug("getting object", zap.String("name", oname))
	ctx := context.Background()
	bucket := s.gcs.Bucket(s.Bucket)
	obj := bucket.Object(oname)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		s.handleError(w, err)
		return
	}

	// write headers
	w.Header().Add("Content-Type", attrs.ContentType)
	w.Header().Add("Content-Length", strconv.FormatInt(attrs.Size, 10))

	// write body
	objr, err := obj.NewReader(ctx)
	if err != nil {
		s.handleError(w, err)
		return
	}
	_, err = io.Copy(w, objr)
	if err != nil {
		s.logger.Warn("cannot write response body", zap.Error(err))
	}
}

func main() {
	// init server
	var err error
	s := Server{}
	pflag.StringVar(&s.Address, "addr", "127.0.0.1:8080", "address to serve")
	pflag.StringVar(&s.Bucket, "bucket", "", "bucket to serve")
	pflag.StringVar(&s.Cred, "cred", "", "path to gcloud credential file")
	pflag.StringVar(&s.Username, "username", "", "username for basic HTTP auth")
	pflag.StringVar(&s.Password, "password", "", "password for basic HTTP auth")
	pflag.BoolVar(&s.Debug, "debug", false, "show debug logs")
	pflag.Parse()
	err = s.Init()
	if err != nil {
		log.Fatalf("cannot init server: %v", err)
	}
	defer func() {
		err := s.logger.Sync()
		if err != nil {
			log.Fatalf("cannot sync logger: %v", err)
		}
	}()

	// serve
	router := httprouter.New()
	router.GET("/*object", s.Handle)
	s.logger.Info("listening", zap.String("addr", s.Address))
	err = http.ListenAndServe(s.Address, router)
	s.logger.Error("server has stopped", zap.Error(err))
}
