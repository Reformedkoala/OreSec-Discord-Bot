package util

type Config struct {
    CTFDToken string `mapstructure:"CTFD_TOKEN"`
    DiscordToken    string `mapstructure:"DISCORD_TOKEN"`
    CTFDAddress  string `mapstructure:"BASE_URL"`
    FirstBloodFile string `mapstructure:"MAP_FILE"`
    FirstBloodChannel string `mapstructure:"BLOOD_CHANNEL"`
    BloodRole string `mapstructure:"BLOOD_ROLE"`
    GuildID string `mapstructure:"GUILD_ID"`
    AppID string `mapstructure:"APP_ID"`
}

type ChallengePost struct {
	Category    string  `json:"category"`
	Decay       int `json:"decay"`
	Description string  `json:"description"`
	Function    string  `json:"function"`
	Initial     int     `json:"initial"`
	Minimum     int     `json:"minimum"`
	Name        string  `json:"name"`
	State       string  `json:"state"`
	Type        string  `json:"type"`
}

type ChallengeSubmit struct {
    Success bool `json:"success"`
    Data ChallengeSubmitData `json:"data"`
}

type ChallengeSubmitData struct {
    Category string `json:"category"` 
    Connection_info string `json:"connection_info"` 
    Decay int `json:"decay"`
    Description string `json:"description"`
    Id int `json:"id"`
    Initial int `json:"initial"`
    Max_attempts int `json:"max_attempts"`
    Minimum int `json:"minimum"`
    Name string `json:"name"`
    Next_id int `json:"next_id"`
    State string `json:"state"`
    Type string `json:"type"`
    Type_data map[string]interface{} `json:"type_data"`
    Value int `json:"value"`
}

type FlagPost struct {
    Challenge_ID int `json:"challenge_id"`
    Content string `json:"content"`
    Data string `json:"data"`
    Type string `json:"type"`
}

type GenericResponse struct {
    Success bool `json:"success"`
    Data map[string]interface{} `json:"data"`
}

type Challenges struct {
    Success bool `json:"success"`
    Data []ChallengeData `json:"data"`
}

type ChallengeData struct {
    Id int `json:"id"` 
    Challenge_type string `json:"type"`
    Name string `json:"name"`
    Value int `json:"values"`
    Solves int `json:"solves"`
    Solved_by_me bool `json:"solved_by_me"`
    Category string `json:"category"`
    Tags []string `json:"tags"`
    Template_page string `json:"template"`
    Script string `json:"script"`
}

type Solves struct {
    Success bool `json:"success"`
    Data []SolveData `json:"data"`
}

type SolveData struct {
    Account_id int `json:"account_id"`
    TeamName string `json:"name"`
    Data string `json:"data"`
    Account_url string `json:"account_url"`
}
