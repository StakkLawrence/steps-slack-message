package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/bitrise-io/go-utils/colorstring"
)

// ConfigsModel ...
type ConfigsModel struct {
	// Slack Inputs
	WebhookURL          string
	Channel             string
	FromUsername        string
	Title        				string
	TitleLink        		string
	Footer			        string
	FiledTitle1	        string
	FiledDetail1        string
	FiledTitle2      	  string
	FiledDetail2        string
	Color               string
	ColorOnError        string
	ThumbURL        		string
	IconURL             string
	// Other Inputs
	IsDebugMode bool
	// Other configs
	IsBuildFailed bool
	IsBuildSucceed bool
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		WebhookURL:          	os.Getenv("webhook_url"),
		Channel:             	os.Getenv("channel"),
		FromUsername:        	os.Getenv("from_username"),
		Title:								os.Getenv("title"),
		TitleLink:						os.Getenv("title_link"),
		Footer:								os.Getenv("footer"),
		FiledTitle1:					os.Getenv("field_title_1"),
		FiledDetail1:					os.Getenv("field_detail_1"),
		FiledTitle2:					os.Getenv("field_title_2"),
		FiledDetail2:					os.Getenv("field_detail_2"),
		Color:               os.Getenv("color"),
		ColorOnError:        os.Getenv("color_on_error"),
		ThumbURL:        	 	 os.Getenv("thumb_url"),
		IconURL:             os.Getenv("icon_url"),
		//
		IsDebugMode: (os.Getenv("is_debug_mode") == "yes"),
		//
		IsBuildFailed: (os.Getenv("STEPLIB_BUILD_STATUS") != "0"),
		IsBuildSucceed: (os.Getenv("STEPLIB_BUILD_STATUS") == "0"),
	}
}

func (configs ConfigsModel) print() {
	fmt.Println("")
	fmt.Println(colorstring.Blue("Slack configs:"))
	fmt.Println(" - WebhookURL:", configs.WebhookURL)
	fmt.Println(" - Channel:", configs.Channel)
	fmt.Println(" - FromUsername:", configs.FromUsername)
	fmt.Println(" - Title:", configs.Title)
	fmt.Println(" - TitleLink:", configs.TitleLink)
	fmt.Println(" - Footer:", configs.Footer)
	fmt.Println(" - FiledTitle1:", configs.FiledTitle1)
	fmt.Println(" - FiledDetail1:", configs.FiledDetail1)
	fmt.Println(" - FiledTitle2:", configs.FiledTitle2)
	fmt.Println(" - FiledDetail2:", configs.FiledDetail2)
	fmt.Println(" - Color:", configs.Color)
	fmt.Println(" - ColorOnError:", configs.ColorOnError)
	fmt.Println(" - ThumbURL:", configs.ThumbURL)
	fmt.Println(" - IconURL:", configs.IconURL)
	fmt.Println("")
	fmt.Println(colorstring.Blue("Other configs:"))
	fmt.Println(" - IsDebugMode:", configs.IsDebugMode)
	fmt.Println(" - IsBuildFailed:", configs.IsBuildFailed)
	fmt.Println(" - IsBuildSucceed:", configs.IsBuildSucceed)
	fmt.Println("")
}

func (configs ConfigsModel) validate() error {
	// required
	if configs.WebhookURL == "" {
		return errors.New("No Webhook URL parameter specified!")
	}
	if configs.Color == "" {
		return errors.New("No Color parameter specified!")
	}

	return nil
}

// FieldModel ...
type AttachmentFieldModel struct {
	Title 			string   `json:"title"`
	Value    		string   `json:"value"`
	Short    		string   `json:"short"`
}

// AttachmentItemModel ...
type AttachmentItemModel struct {
	Fallback 			string   `json:"fallback"`
	Color    			string   `json:"color,omitempty"`
	Title    			string   `json:"title"`
	TitleLink     string   `json:"title_link"`
	Text     			string   `json:"text"`
	Fields []AttachmentFieldModel `json:"fields,omitempty"`
	ThumbURL  		string `json:"thumb_url,omitempty"`
	Footer  			string `json:"footer,omitempty"`
}

// RequestParams ...
type RequestParams struct {
	// - required
	Text string `json:"text"`
	// OR use attachment instead of text, for better formatting
	Attachments []AttachmentItemModel `json:"attachments,omitempty"`
	// - optional
	Channel   *string `json:"channel"`
	Username  *string `json:"username"`
	EmojiIcon *string `json:"icon_emoji"`
	IconURL   *string `json:"icon_url"`
}

// CreatePayloadParam ...
func CreatePayloadParam(configs ConfigsModel) (string, error) {
	// - required
	msgColor := configs.Color
	if configs.IsBuildFailed {
		if configs.ColorOnError == "" {
			fmt.Println(colorstring.Yellow(" (i) Build failed but no color_on_error defined, using default."))
		} else {
			msgColor = configs.ColorOnError
		}
	}

	// msgText := configs.Message
	// if configs.IsBuildFailed {
	// 	if configs.MessageOnError == "" {
	// 		fmt.Println(colorstring.Yellow(" (i) Build failed but no message_on_error defined, using default."))
	// 	} else {
	// 		msgText = configs.MessageOnError
	// 	}
	// }

	fieldTitle1 := configs.FiledTitle1
	fieldDetail1 := configs.FiledDetail1
	fieldTitle2 := configs.FiledTitle2
	fieldDetail2 := configs.FiledDetail2

	fields := []AttachmentFieldModel{}

	if configs.IsBuildSucceed {
		if fieldTitle1 != "" && fieldTitle2 != "" {
			fields = []AttachmentFieldModel{
				{
					Title:  fieldTitle1,
					Value:  fieldDetail1,
				},
				{
					Title: 	fieldTitle2,
					Value:  fieldDetail2,
				},
			}
		} else if fieldTitle1 != "" {
			fields = []AttachmentFieldModel{
				{
					Title: 	fieldTitle1,
					Value:  fieldDetail1,
				},
			}
		}
	}

	reqParams := RequestParams{
		Attachments: []AttachmentItemModel{
			{
				Fallback: 	configs.Title,
				Color:    	msgColor,
				Title: 			configs.Title,
				TitleLink:	configs.TitleLink,
				Fields: 		fields,
				ThumbURL:   configs.ThumbURL,
				Footer:			configs.Footer,
			},
		},
	}

	// - optional
	reqChannel := configs.Channel
	if reqChannel != "" {
		reqParams.Channel = &reqChannel
	}

	reqUsername := configs.FromUsername
	if reqUsername != "" {
		reqParams.Username = &reqUsername
	}

	reqIconURL := configs.IconURL
	if reqIconURL != "" {
		reqParams.IconURL = &reqIconURL
	}

	// if Icon URL defined ignore the emoji input
	if reqParams.IconURL != nil {
		reqParams.EmojiIcon = nil
	}

	if configs.IsDebugMode {
		fmt.Printf("Parameters: %#v\n", reqParams)
	}

	// JSON serialize the request params
	reqParamsJSONBytes, err := json.Marshal(reqParams)
	if err != nil {
		return "", nil
	}
	reqParamsJSONString := string(reqParamsJSONBytes)

	return reqParamsJSONString, nil
}

func main() {
	configs := createConfigsModelFromEnvs()
	configs.print()
	if err := configs.validate(); err != nil {
		fmt.Println()
		fmt.Println(colorstring.Red("Issue with input:"), err)
		fmt.Println()
		os.Exit(1)
	}

	//
	// request URL
	requestURL := configs.WebhookURL

	//
	// request parameters
	reqParamsJSONString, err := CreatePayloadParam(configs)
	if err != nil {
		fmt.Println(colorstring.Red("Failed to create JSON payload:"), err)
		os.Exit(1)
	}
	if configs.IsDebugMode {
		fmt.Println()
		fmt.Println("JSON payload: ", reqParamsJSONString)
	}

	//
	// send request
	resp, err := http.PostForm(requestURL,
		url.Values{"payload": []string{reqParamsJSONString}})
	if err != nil {
		fmt.Println(colorstring.Red("Failed to send the request:"), err)
		os.Exit(1)
	}

	//
	// process the response
	body, err := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println()
		fmt.Println(colorstring.Red("Request failed"))
		fmt.Println("Response from Slack: ", bodyStr)
		fmt.Println()
		os.Exit(1)
	}

	if configs.IsDebugMode {
		fmt.Println()
		fmt.Println("Response from Slack: ", bodyStr)
	}
	fmt.Println()
	fmt.Println(colorstring.Green("Slack message successfully sent! 🚀"))
	fmt.Println()
	os.Exit(0)
}
