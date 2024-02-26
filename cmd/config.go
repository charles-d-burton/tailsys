package cmd

//TODO: Figure this out for load/save of config file

// import (
// 	"os"
// 	"path/filepath"

// 	"github.com/kirsle/configdir"
// 	"gopkg.in/ini.v1"
// )

// type Client struct {
//   CoordinationServerTags string `ini:"coordination_server_tags"`
// }

// type Auth struct {
//   ClientID string `ini:"client_id"`
//   ClientSecret string `ini:"client_secret"`
//   AuthKey string `ini:"auth_key"`
// }

// type Config struct {
//   TailsysKey string `ini:"tailsys_key"`
//   Port int `ini:"port"`
//   Client
//   Auth
// }

// func (config *Config) Load() error {
//   configPath := configdir.LocalConfig("tailsys")
//   err := configdir.MakePath(configPath)
//   if err != nil {
//     return err
//   }
//   configFile := filepath.Join(configPath, "tailsys.toml")

//   //Config file exists?
//   if _, err = os.Stat(configFile); os.IsNotExist(err) {
//     _, err := os.Create(configFile) 
//     if err != nil {
//       return err
//     }
//   }

//   _, err = ini.Load(configFile)
//   if err != nil {
//     return err
//   }

//   return nil
// }

// func (config *Config) Valid() bool {
//   return true
// }
