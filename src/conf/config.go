package conf

import "github.com/spf13/viper"

var s *viper.Viper
var sw *viper.Viper
var m *viper.Viper

func init() {
	s = viper.New()
	s.SetConfigType("yaml")
	s.SetConfigName("slackbot")
	s.AddConfigPath("conf/environments/")

	sw = viper.New()
	sw.SetConfigType("yaml")
	sw.SetConfigName("stay_watch")
	sw.AddConfigPath("conf/environments/")

	m = viper.New()
	m.SetConfigType("yaml")
	m.SetConfigName("mysql")
	m.AddConfigPath("conf/environments/")
}

func GetSlackConfig() *viper.Viper {
	if err := s.ReadInConfig(); err != nil {
		return nil
	}
	return s
}

func GetStayWatchConfig() *viper.Viper {
	if err := sw.ReadInConfig(); err != nil {
		return nil
	}
	return s
}

func GetMysqlConfig() *viper.Viper {
	if err := m.ReadInConfig(); err != nil {
		return nil
	}
	return m
}