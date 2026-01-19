package install

import (
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
	GetUseDefault(&iarg)
	GetConfigPath(&iarg)
	if iarg.UseDefaultConfig {
		iarg.Config = config.DefaultConfig()
	} else {
		iarg.Config = config.Config{}
		GetEnvironments(&iarg)
		GetPorts(&iarg)
	}
	GetDbDriver(&iarg)
	switch iarg.DB_Driver {
	case MYSQL, POSTGRES:
		GetFullSqlSetup(&iarg)
	default:
		GetLiteSqlSetup(&iarg)
	}
	GetBuckets(&iarg)

	return &iarg, nil
}

func GetUseDefault(i *InstallArguments) {
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
		utility.DefaultLogger.Fatal("", err)
	}

	i.UseDefaultConfig = useDefault

}
func GetConfigPath(i *InstallArguments) {
	path := "config.json"
	f := huh.NewInput().Title("Where would you like to save the config?").
		Value(&path)
	err := f.Run()

	if err != nil {
		utility.DefaultLogger.Fatal("", err)
	}

	i.ConfigPath = path
}
func GetEnvironments(i *InstallArguments) {
	environments := map[string]string{}
	devUrl := "localhost"
	stageUrl := "localhost"
	prodUrl := "localhost"
	f1 := huh.NewInput().Title("Development URL").Value(&devUrl)
	f2 := huh.NewInput().Title("Stage URL").Value(&stageUrl)
	f3 := huh.NewInput().Title("Production URL").Value(&prodUrl)
	g := huh.NewGroup(f1, f2, f3)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		utility.DefaultLogger.Fatal("", err)
	}
	environments["development"] = devUrl
	environments["staging"] = stageUrl
	environments["prodUrl"] = prodUrl
	i.Config.Environment_Hosts = environments
}

func GetPorts(i *InstallArguments) {
	httpPort := "1234"
	httpsPort := "4000"
	sshPort := "2233"
	f1 := huh.NewInput().Title("http port").Value(&httpPort)
	f2 := huh.NewInput().Title("https port").Value(&httpsPort)
	f3 := huh.NewInput().Title("ssh port").Value(&sshPort)
	g := huh.NewGroup(f1, f2, f3)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		utility.DefaultLogger.Fatal("", err)
	}
    i.Config.Port = httpPort
    i.Config.SSL_Port = httpsPort
    i.Config.SSH_Port = sshPort
}

func GetDbDriver(i *InstallArguments) {
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
		utility.DefaultLogger.Error("", err)
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
}

func GetFullSqlSetup(i *InstallArguments) {
	fUrl := "localhost"
	fName := "modula_db"
	fUser := "modula_db_u"
	fPassword, err := utility.MakeRandomString()
	if err != nil {
		utility.DefaultLogger.Fatal("Failed to generate random password", err)
	}
	f1 := huh.NewInput().Title("URL to Database Host: ").Value(&fUrl)
	f2 := huh.NewInput().Title("Database Name: ").Value(&fName)
	f3 := huh.NewInput().Title("Database User: ").Value(&fUser)
	f4 := huh.NewInput().Title("Database Password: ").Value(&fPassword)
	g := huh.NewGroup(f1, f2, f3, f4)
	f := huh.NewForm(g)
	err = f.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	i.Config.Db_URL = fUrl
	i.Config.Db_Name = fName
	i.Config.Db_User = fUser
	i.Config.Db_Password = fPassword
}

func GetLiteSqlSetup(i *InstallArguments) {
	fUrl := "modula.db"
	fName := "modula_db"
	f1 := huh.NewInput().Title("URL to Database Host: ").Value(&fUrl)
	f2 := huh.NewInput().Title("Database Name: ").Value(&fName)
	g := huh.NewGroup(f1, f2)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	i.Config.Db_URL = fUrl
	i.Config.Db_Name = fName

}

func GetBuckets(i *InstallArguments) {
	bAccess := ""
	bSecret := ""
	bRegion := ""
	bEndpoint := ""
	bMedia := ""
	bBackup := ""
	f1 := huh.NewInput().Title("S3 Bucket Access Token").Value(&bAccess)
	f2 := huh.NewInput().Title("S3 Bucket Secret Token").Value(&bSecret)
	f3 := huh.NewInput().Title("S3 Bucket Region").Value(&bRegion)
	f4 := huh.NewInput().Title("S3 Bucket Endpoint").Value(&bEndpoint)
	f5 := huh.NewInput().Title("S3 Bucket Media Path").Value(&bMedia)
	f6 := huh.NewInput().Title("S3 Bucket Backup Path").Value(&bBackup)
	err := f1.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	err = f2.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	err = f3.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	err = f4.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	err = f5.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	err = f6.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	i.Config.Bucket_Access_Key = bAccess
	i.Config.Bucket_Secret_Key = bSecret
	i.Config.Bucket_Region = bRegion
	i.Config.Bucket_Endpoint = bEndpoint
	i.Config.Bucket_Media = bMedia
	i.Config.Bucket_Backup = bBackup
}
