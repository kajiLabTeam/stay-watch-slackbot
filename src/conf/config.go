package conf

import "github.com/spf13/viper"

var s *viper.Viper

func init() {
	s = viper.New()
	s.SetConfigType("yaml")
	s.SetConfigName("slackbot")
	s.AddConfigPath("conf/environments/")
}

func GetSlackConfig() *viper.Viper {
	if err := s.ReadInConfig(); err != nil {
		return nil
	}
	return s
}
