package importer

type Config struct {
	GetTokenFunctionName string   `mapstructure:"get_token_function_name"`
	EventsTableName      string   `mapstructure:"events_table_name"`
	ImporterLogTableName string   `mapstructure:"importer_log_table_name"`
	MeetupGroupNames     []string `mapstructure:"meetup_group_names"`
}
