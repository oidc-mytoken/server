package dbcl

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
)

// RunDBCommands executes SQL stmts through the mysql cli
func RunDBCommands(cmds string, dbConfig config.DBConf, printOutput bool) error {
	mysqlCmd := fmt.Sprintf("mysql -u%s -p%s --protocol tcp -h %s",
		dbConfig.User, dbConfig.GetPassword(), dbConfig.Hosts[0])
	if dbConfig.DB != "" {
		mysqlCmd += fmt.Sprintf(" %s", dbConfig.DB)
	}
	cmd := exec.Command("sh", "-c", mysqlCmd)
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	if printOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if _, err = cmdIn.Write([]byte(cmds)); err != nil {
		return errors.WithStack(err)
	}
	if err = cmdIn.Close(); err != nil {
		return errors.WithStack(err)
	}
	return cmd.Run()
}
