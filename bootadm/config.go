// +build solaris

package bootadm

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"git.wegmueller.it/opencloud/opencloud/uname"
	"github.com/toasterson/mozaik/logger"
)

const (
	BootLoaderTypeLoader = 0
	BootLoaderTypeGrub   = 1
)

const loaderConfFile = "/%s/boot/menu.lst"

const loaderBootConfig = `title {{.BEName}}
bootfs {{.Rpool}}/ROOT/{{.BEName}}`

const grubBootConfig = `default 0
timeout 3
title {{.BEName}}
findroot (pool_{{.Rpool}},0,a)
bootfs {{.Rpool}}/ROOT/{{.BEName}}
kernel$ /platform/i86pc/kernel/$ISADIR/unix -B $ZFS-BOOTFS
module$ /platform/i86pc/$ISADIR/boot_archive`

const grubConfFile = "/%s/boot/grub/menu.lst"

const xenBootConfig = `default 0
timeout 3
title {{.BEName}}
findroot (pool_{{.Rpool}},1,a)
bootfs {{.Rpool}}/ROOT/{{.BEName}}
kernel$ /platform/i86pc/kernel/amd64/unix -B $ZFS-BOOTFS
module$ /platform/i86pc/amd64/boot_archive`

type loaderType int

type BootConfig struct {
	Type        loaderType
	RPoolName   string
	BEName      string
	BootOptions []string //TODO Implement
}

func CreateBootConfigurationFiles(rootDir string, conf BootConfig) (err error) {
	if rootDir == "" {
		rootDir = "/"
	}

	hplatform := uname.GetHardwarePlatform()
	config := loaderBootConfig
	confLocation := loaderConfFile
	if hplatform == uname.HardwarePlatformXen {
		config = xenBootConfig
		confLocation = grubConfFile
		logger.Info("Configuring Bootloader for Xen")
	} else if conf.Type == BootLoaderTypeGrub {
		config = grubBootConfig
		confLocation = grubConfFile
		logger.Info("Using Grub Configuration for Installation")
	}
	tmplConfig, err := template.New("BootConfig").Parse(config)
	if err != nil {
		return
	}
	var out bytes.Buffer
	err = tmplConfig.Execute(&out, conf)
	if err != nil {
		return
	}
	if err = os.Mkdir(fmt.Sprintf("/%s/boot", conf.RPoolName), os.ModeDir); err != nil {
		return
	}
	confFile, err := os.Create(fmt.Sprintf(confLocation, conf.RPoolName))
	if err != nil {
		return
	}
	logger.Info("Writing Configuration")
	logger.Trace(out.String())
	_, err = confFile.Write(out.Bytes())
	return
}
