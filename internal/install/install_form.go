package install

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

type DatabaseDriver string

const (
	SQLITE   DatabaseDriver = "sqlite"
	MYSQL    DatabaseDriver = "mysql"
	POSTGRES DatabaseDriver = "postgres"
)

type InstallArguments struct {
	UseDefaultConfig bool           `json:"use_default_config"`
	ConfigPath       string         `json:"config_path"`
	Config           config.Config  `json:"config"`
	DB_Driver        DatabaseDriver `json:"db_driver"`
	Create_Tables    bool           `json:"create_tables"`
}

func RunInstallIO() (*InstallArguments, error) {
	iarg := InstallArguments{}

	if err := GetUseDefault(&iarg); err != nil {
		return nil, err
	}
	if err := GetConfigPath(&iarg); err != nil {
		return nil, err
	}

	if iarg.UseDefaultConfig {
		iarg.Config = config.DefaultConfig()
	} else {
		iarg.Config = config.Config{}
		if err := GetEnvironments(&iarg); err != nil {
			return nil, err
		}
		if err := GetPorts(&iarg); err != nil {
			return nil, err
		}
	}

	if err := GetDbDriver(&iarg); err != nil {
		return nil, err
	}

	switch iarg.DB_Driver {
	case MYSQL, POSTGRES:
		if err := GetFullSqlSetup(&iarg); err != nil {
			return nil, err
		}
	default:
		if err := GetLiteSqlSetup(&iarg); err != nil {
			return nil, err
		}
	}

	if err := GetBuckets(&iarg); err != nil {
		return nil, err
	}

	return &iarg, nil
}

func GetUseDefault(i *InstallArguments) error {
	var useDefault bool
	f := huh.NewSelect[bool]()
	f.Options(huh.Option[bool]{
		Key:   "Yes",
		Value: true,
	}, huh.Option[bool]{
		Key:   "No",
		Value: false,
	}).Title("Would you like to use the default configuration?").
		Value(&useDefault)

	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.UseDefaultConfig = useDefault
	return nil
}

func GetConfigPath(i *InstallArguments) error {
	path := "config.json"
	f := huh.NewInput().
		Title("Where would you like to save the config?").
		Value(&path).
		Validate(ValidateConfigPath)

	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.ConfigPath = path
	return nil
}

func GetEnvironments(i *InstallArguments) error {
	environments := map[string]string{}
	devUrl := "localhost"
	stageUrl := "localhost"
	prodUrl := "localhost"

	f1 := huh.NewInput().Title("Development URL").Value(&devUrl).Validate(ValidateURL)
	f2 := huh.NewInput().Title("Stage URL").Value(&stageUrl).Validate(ValidateURL)
	f3 := huh.NewInput().Title("Production URL").Value(&prodUrl).Validate(ValidateURL)

	g := huh.NewGroup(f1, f2, f3)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	environments["development"] = devUrl
	environments["staging"] = stageUrl
	environments["production"] = prodUrl
	i.Config.Environment_Hosts = environments
	return nil
}

func GetPorts(i *InstallArguments) error {
	httpPort := "1234"
	httpsPort := "4000"
	sshPort := "2233"

	f1 := huh.NewInput().Title("HTTP port").Value(&httpPort).Validate(ValidatePort)
	f2 := huh.NewInput().Title("HTTPS port").Value(&httpsPort).Validate(ValidatePort)
	f3 := huh.NewInput().Title("SSH port").Value(&sshPort).Validate(ValidatePort)

	g := huh.NewGroup(f1, f2, f3)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Port = httpPort
	i.Config.SSL_Port = httpsPort
	i.Config.SSH_Port = sshPort
	return nil
}

func GetDbDriver(i *InstallArguments) error {
	var driver rune
	f := huh.NewSelect[rune]()
	f.Options(huh.Option[rune]{
		Key:   "Sqlite",
		Value: '1',
	}, huh.Option[rune]{
		Key:   "MySql",
		Value: '2',
	}, huh.Option[rune]{
		Key:   "Postgres",
		Value: '3',
	}).Title("What Database do you want to use?").Value(&driver)

	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	switch driver {
	case '1':
		i.DB_Driver = SQLITE
		i.Config.Db_Driver = config.DbDriver(SQLITE)
	case '2':
		i.DB_Driver = MYSQL
		i.Config.Db_Driver = config.DbDriver(MYSQL)
	case '3':
		i.DB_Driver = POSTGRES
		i.Config.Db_Driver = config.DbDriver(POSTGRES)
	}
	return nil
}

func GetFullSqlSetup(i *InstallArguments) error {
	fUrl := "localhost"
	fName := "modula_db"
	fUser := "modula_db_u"
	fPassword, err := utility.MakeRandomString()
	if err != nil {
		return err
	}

	f1 := huh.NewInput().Title("URL to Database Host").Value(&fUrl).Validate(ValidateURL)
	f2 := huh.NewInput().Title("Database Name").Value(&fName).Validate(ValidateDBName)
	f3 := huh.NewInput().Title("Database User").Value(&fUser).Validate(ValidateNotEmpty("database user"))
	f4 := huh.NewInput().Title("Database Password").Value(&fPassword).EchoMode(huh.EchoModePassword)

	g := huh.NewGroup(f1, f2, f3, f4)
	f := huh.NewForm(g)
	err = f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Db_URL = fUrl
	i.Config.Db_Name = fName
	i.Config.Db_User = fUser
	i.Config.Db_Password = fPassword
	return nil
}

func GetLiteSqlSetup(i *InstallArguments) error {
	fUrl := "modula.db"
	fName := "modula_db"

	f1 := huh.NewInput().Title("Database file path").Value(&fUrl).Validate(ValidateDBPath)
	f2 := huh.NewInput().Title("Database Name").Value(&fName).Validate(ValidateDBName)

	g := huh.NewGroup(f1, f2)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Db_URL = fUrl
	i.Config.Db_Name = fName
	return nil
}

func GetBuckets(i *InstallArguments) error {
	bAccess := ""
	bSecret := ""
	bRegion := ""
	bEndpoint := ""
	bMedia := ""
	bBackup := ""

	f1 := huh.NewInput().Title("S3 Bucket Access Key (leave empty to skip)").Value(&bAccess)
	f2 := huh.NewInput().Title("S3 Bucket Secret Key").Value(&bSecret).EchoMode(huh.EchoModePassword)
	f3 := huh.NewInput().Title("S3 Bucket Region").Value(&bRegion)
	f4 := huh.NewInput().Title("S3 Bucket Endpoint").Value(&bEndpoint)
	f5 := huh.NewInput().Title("S3 Bucket Media Path").Value(&bMedia)
	f6 := huh.NewInput().Title("S3 Bucket Backup Path").Value(&bBackup)

	g := huh.NewGroup(f1, f2, f3, f4, f5, f6)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Bucket_Access_Key = bAccess
	i.Config.Bucket_Secret_Key = bSecret
	i.Config.Bucket_Region = bRegion
	i.Config.Bucket_Endpoint = bEndpoint
	i.Config.Bucket_Media = bMedia
	i.Config.Bucket_Backup = bBackup
	return nil
}
