package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/config"
	install "github.com/hegner123/modulacms/internal/install"
)

func StatusBlock() string {
	verbose := false
	c := config.Env
	bucketStatus, _ := install.CheckBucket(&verbose)
	dbStatus, _ := install.CheckDb(&verbose, c)
	adminSite := "Admin Site: "
	clientSite := "Client Site: "
	sshConnection := "SSH: "
	bucketConnection := "Bucket Status:"
	dbDriver := "Driver: "
	dbUrl := "URL: "
	dbDriverStatus := dbStatus.Driver
	dbUrlStatus := dbStatus.URL

	if dbStatus.Err != nil {
		dbDriver = "Error"
		dbDriverStatus = dbStatus.Err.Error()
		dbUrl = ""
		dbUrlStatus = ""

	}

	labels := lipgloss.JoinVertical(
		lipgloss.Top,
		adminSite,
		clientSite,
		sshConnection,
		bucketConnection,
		"DB",
		dbDriver,
		dbUrl,
	)
	values := lipgloss.JoinVertical(
		lipgloss.Top,
		c.Admin_Site,
		c.Client_Site,
		fmt.Sprint(c.SSH_Host, ":", c.SSH_Port),
		bucketStatus,
		"",
		dbDriverStatus,
		dbUrlStatus,
	)
    vStyle := lipgloss.NewStyle().MarginLeft(4)
	v := lipgloss.JoinHorizontal(lipgloss.Center, labels, vStyle.Render(values))
	return v
}
