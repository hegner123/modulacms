package install

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/auth"
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
	UseDefaultConfig  bool           `json:"use_default_config"`
	ConfigPath        string         `json:"config_path"`
	Config            config.Config  `json:"config"`
	DB_Driver         DatabaseDriver `json:"db_driver"`
	Create_Tables     bool           `json:"create_tables"`
	AdminPasswordHash string         `json:"-"`
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
		if err := GetDomains(&iarg); err != nil {
			return nil, err
		}
		if err := GetCORS(&iarg); err != nil {
			return nil, err
		}
		if err := GetCertDir(&iarg); err != nil {
			return nil, err
		}
		if err := GetCookie(&iarg); err != nil {
			return nil, err
		}
		if err := GetOutputFormat(&iarg); err != nil {
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

	if err := GetOAuthOptional(&iarg); err != nil {
		return nil, err
	}

	if err := GetAdminPassword(&iarg); err != nil {
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
	bForcePathStyle := true

	f1 := huh.NewInput().Title("S3 Bucket Access Key (leave empty to skip)").Value(&bAccess)
	f2 := huh.NewInput().Title("S3 Bucket Secret Key").Value(&bSecret).EchoMode(huh.EchoModePassword)
	f3 := huh.NewInput().Title("S3 Bucket Region").Value(&bRegion)
	f4 := huh.NewInput().Title("S3 Bucket Endpoint").Value(&bEndpoint)
	f5 := huh.NewInput().Title("S3 Bucket Media Path").Value(&bMedia)
	f6 := huh.NewInput().Title("S3 Bucket Backup Path").Value(&bBackup)
	f7 := huh.NewConfirm().Title("Force S3 path-style addressing? (required for MinIO/Linode)").Value(&bForcePathStyle)

	g := huh.NewGroup(f1, f2, f3, f4, f5, f6, f7)
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
	i.Config.Bucket_Force_Path_Style = bForcePathStyle
	return nil
}

func GetDomains(i *InstallArguments) error {
	clientSite := "localhost"
	adminSite := "localhost"

	f1 := huh.NewInput().Title("Client site domain").Value(&clientSite).Validate(ValidateURL)
	f2 := huh.NewInput().Title("Admin site domain").Value(&adminSite).Validate(ValidateURL)

	g := huh.NewGroup(f1, f2)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Client_Site = clientSite
	i.Config.Admin_Site = adminSite
	return nil
}

func GetCORS(i *InstallArguments) error {
	origins := "http://localhost:3000"
	credentials := false

	f1 := huh.NewInput().Title("CORS allowed origins (comma-separated)").Value(&origins)
	f2 := huh.NewConfirm().Title("Allow CORS credentials?").Value(&credentials)

	g := huh.NewGroup(f1, f2)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	parsed := strings.Split(origins, ",")
	for idx, o := range parsed {
		parsed[idx] = strings.TrimSpace(o)
	}

	i.Config.Cors_Origins = parsed
	i.Config.Cors_Credentials = credentials
	i.Config.Cors_Methods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	i.Config.Cors_Headers = []string{"Content-Type", "Authorization"}
	return nil
}

func GetCertDir(i *InstallArguments) error {
	certDir := "./"

	f := huh.NewInput().
		Title("Certificate directory path").
		Value(&certDir).
		Validate(ValidateDirPath)

	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Cert_Dir = certDir
	return nil
}

func GetCookie(i *InstallArguments) error {
	cookieName := "modula_cms"

	f := huh.NewInput().
		Title("Cookie name").
		Value(&cookieName).
		Validate(ValidateCookieName)

	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Cookie_Name = cookieName
	return nil
}

func GetOutputFormat(i *InstallArguments) error {
	var format string
	f := huh.NewSelect[string]().
		Options(
			huh.NewOption("Raw (default)", "raw"),
			huh.NewOption("Clean", "clean"),
			huh.NewOption("Contentful", "contentful"),
			huh.NewOption("Sanity", "sanity"),
			huh.NewOption("Strapi", "strapi"),
			huh.NewOption("WordPress", "wordpress"),
		).
		Title("Output format").
		Value(&format)

	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	i.Config.Output_Format = config.OutputFormat(format)
	return nil
}

func GetOAuthOptional(i *InstallArguments) error {
	configureOAuth := false
	f := huh.NewConfirm().
		Title("Would you like to configure OAuth?").
		Value(&configureOAuth)

	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	if !configureOAuth {
		i.Config.Oauth_Client_Id = ""
		i.Config.Oauth_Client_Secret = ""
		i.Config.Oauth_Scopes = []string{}
		i.Config.Oauth_Endpoint = map[config.Endpoint]string{}
		i.Config.Oauth_Provider_Name = ""
		i.Config.Oauth_Redirect_URL = ""
		i.Config.Oauth_Success_Redirect = ""
		return nil
	}

	return GetOAuth(i)
}

// GetAdminPassword prompts for the system admin password with confirmation,
// validates it, and stores the bcrypt hash in InstallArguments.
func GetAdminPassword(i *InstallArguments) error {
	password := ""
	confirm := ""

	f1 := huh.NewInput().
		Title("System admin password (min 8 characters)").
		Value(&password).
		EchoMode(huh.EchoModePassword).
		Validate(ValidatePassword)

	f2 := huh.NewInput().
		Title("Confirm admin password").
		Value(&confirm).
		EchoMode(huh.EchoModePassword).
		Validate(ValidatePassword)

	g := huh.NewGroup(f1, f2)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	i.AdminPasswordHash = hash
	return nil
}

func GetOAuth(i *InstallArguments) error {
	providerName := ""
	clientID := ""
	clientSecret := ""
	authURL := ""
	tokenURL := ""
	userInfoURL := ""
	redirectURL := "http://localhost:8080/auth/callback"
	scopes := "openid,profile,email"
	successRedirect := "/"

	f1 := huh.NewInput().Title("OAuth provider name").Value(&providerName).Validate(ValidateNotEmpty("provider name"))
	f2 := huh.NewInput().Title("OAuth client ID").Value(&clientID).Validate(ValidateNotEmpty("client ID"))
	f3 := huh.NewInput().Title("OAuth client secret").Value(&clientSecret).EchoMode(huh.EchoModePassword)
	f4 := huh.NewInput().Title("OAuth authorization URL").Value(&authURL).Validate(ValidateURL)
	f5 := huh.NewInput().Title("OAuth token URL").Value(&tokenURL).Validate(ValidateURL)
	f6 := huh.NewInput().Title("OAuth user info URL").Value(&userInfoURL).Validate(ValidateURL)
	f7 := huh.NewInput().Title("OAuth redirect URL").Value(&redirectURL).Validate(ValidateURL)
	f8 := huh.NewInput().Title("OAuth scopes (comma-separated)").Value(&scopes)
	f9 := huh.NewInput().Title("OAuth success redirect path").Value(&successRedirect)

	g := huh.NewGroup(f1, f2, f3, f4, f5, f6, f7, f8, f9)
	f := huh.NewForm(g)
	err := f.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserAborted()
		}
		return err
	}

	parsedScopes := strings.Split(scopes, ",")
	for idx, s := range parsedScopes {
		parsedScopes[idx] = strings.TrimSpace(s)
	}

	i.Config.Oauth_Provider_Name = providerName
	i.Config.Oauth_Client_Id = clientID
	i.Config.Oauth_Client_Secret = clientSecret
	i.Config.Oauth_Endpoint = map[config.Endpoint]string{
		config.OauthAuthURL:     authURL,
		config.OauthTokenURL:    tokenURL,
		config.OauthUserInfoURL: userInfoURL,
	}
	i.Config.Oauth_Redirect_URL = redirectURL
	i.Config.Oauth_Scopes = parsedScopes
	i.Config.Oauth_Success_Redirect = successRedirect
	return nil
}
