package util

type Config struct {
    CTFDToken string `mapstructure:"CTFD_TOKEN"`
    DiscordToken    string `mapstructure:"DISCORD_TOKEN"`
    CTFDAddress  string `mapstructure:"BASE_URL"`
    FirstBloodFile string `mapstructure:"MAP_FILE"`
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
