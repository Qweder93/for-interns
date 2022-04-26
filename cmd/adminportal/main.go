// Copyright (C) 2021 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"cleanmasters/adminportal/managers"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/spf13/cobra"
	"github.com/zeebo/errs"

	"cleanmasters"
	"cleanmasters/database"
	"cleanmasters/internal/logger/zaplog"
)

// Config defines cleanmansters configuration.
type Config struct {
	Database string `help:"cleanmasters database connection string" releaseDefault:"postgres://" devDefault:"postgres://"`

	cleanmasters.Config
}

var (
	rootCmd = &cobra.Command{
		Use:   "cleanmasters-admin",
		Short: "Cleanmasters admin panel",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Runs the cleanmasters admin panel",
		RunE:  cmdRun,
	}
	createSchemaCmd = &cobra.Command{
		Use:   "create-schema",
		Short: "Create schemas for cleanmasters databases",
		RunE:  cmdCreateSchema,
	}
	setupCmd = &cobra.Command{
		Use:         "setup",
		Short:       "Create config files",
		RunE:        cmdSetup,
		Annotations: map[string]string{"type": "setup"},
	}

	runCfg           Config
	setupCfg         Config
	defaultConfigDir = applicationDir("cleanmasters", "admin")
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(createSchemaCmd)
}

func cmdSetup(cmd *cobra.Command, args []string) (err error) {
	log := zaplog.NewLog()

	setupDir, err := filepath.Abs(defaultConfigDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(setupDir, os.ModePerm)
	if err != nil {
		return err
	}

	configFile, err := os.Create(path.Join(setupDir, "config.json"))
	if err != nil {
		log.Error("could not create config file", err)
		return err
	}

	defer func() {
		err = errs.Combine(err, configFile.Close())
	}()

	jsonData, err := json.MarshalIndent(setupCfg, "", "    ")
	if err != nil {
		log.Error("could not marshal config", err)
		return err
	}

	_, err = configFile.Write(jsonData)
	if err != nil {
		log.Error("could not write to config", err)
		return err
	}

	return nil
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	log := zaplog.NewLog()

	runCfg, err = readConfig()
	if err != nil {
		log.Error("Could not read config from default place", err)
		return err
	}

	db, err := database.Open(runCfg.Database)
	if err != nil {
		log.Error("Error starting master database on cleanmasters admin panel", err)
		return err
	}

	defer func() {
		err = errs.Combine(err, db.Close())
	}()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("qwe"), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error hashing", err)
	}

	err = db.Managers().Add(ctx, managers.Manager{
		ID:           uuid.New(),
		FirstName:    "Chechen",
		LastName:     "Alan",
		Email:        "qwe@ukr.net",
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		log.Error("Error creating manager", err)
	}

	peer, err := cleanmasters.NewPeer(log, db, runCfg.Config)
	if err != nil {
		log.Error("Error starting cleanmasters admin panel", err)
		return err
	}

	runError := peer.RunAdmin(ctx)
	closeError := peer.Close()

	return errs.Combine(runError, closeError)
}

func cmdCreateSchema(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	log := zaplog.NewLog()

	runCfg, err = readConfig()
	if err != nil {
		log.Error("Could not read config from default place", err)
		return err
	}

	db, err := database.Open(runCfg.Database)
	if err != nil {
		return errs.New("error connecting to master database on cleanmasters admin panel: %+v", err)
	}
	defer func() {
		err = errs.Combine(err, db.Close())
	}()

	err = db.CreateSchema(ctx)
	if err != nil {
		log.Error("error creating schema", err)
		return errs.New("error creating database schemas for cleanmasters db: %+v", err)
	}

	return nil
}

// applicationDir returns best base directory for specific OS.
func applicationDir(subdir ...string) string {
	for i := range subdir {
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			subdir[i] = strings.Title(subdir[i])
		} else {
			subdir[i] = strings.ToLower(subdir[i])
		}
	}

	var appdir string

	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "windows":
		// Windows standards: https://msdn.microsoft.com/en-us/library/windows/apps/hh465094.aspx?f=255&MSPPError=-2147217396
		for _, env := range []string{"AppData", "AppDataLocal", "UserProfile", "Home"} {
			val := os.Getenv(env)
			if val != "" {
				appdir = val
				break
			}
		}
	case "darwin":
		// Mac standards: https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/MacOSXDirectories/MacOSXDirectories.html
		appdir = filepath.Join(home, "Library", "Application Support")
	case "linux":
		fallthrough
	default:
		// Linux standards: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
		appdir = os.Getenv("XDG_DATA_HOME")
		if appdir == "" && home != "" {
			appdir = filepath.Join(home, ".local", "share")
		}
	}

	return filepath.Join(append([]string{appdir}, subdir...)...)
}

// readConfig reads config from default config dir.
func readConfig() (config Config, err error) {
	configBytes, err := ioutil.ReadFile(path.Join(defaultConfigDir, "config.json"))
	if err != nil {
		return Config{}, err
	}

	return config, json.Unmarshal(configBytes, &config)
}
