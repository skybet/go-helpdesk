package slack

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

const (
	DEFAULT_USER_PHOTO_CROP_X = -1
	DEFAULT_USER_PHOTO_CROP_Y = -1
	DEFAULT_USER_PHOTO_CROP_W = -1
)

// UserProfile contains all the information details of a given user
type UserProfile struct {
	AvatarHash            string                  `json:"avatar_hash"`
	FirstName             string                  `json:"first_name"`
	LastName              string                  `json:"last_name"`
	RealName              string                  `json:"real_name"`
	RealNameNormalized    string                  `json:"real_name_normalized"`
	DisplayName           string                  `json:"display_name"`
	DisplayNameNormalized string                  `json:"display_name_normalized"`
	Email                 string                  `json:"email"`
	Skype                 string                  `json:"skype"`
	Phone                 string                  `json:"phone"`
	Image24               string                  `json:"image_24"`
	Image32               string                  `json:"image_32"`
	Image48               string                  `json:"image_48"`
	Image72               string                  `json:"image_72"`
	Image192              string                  `json:"image_192"`
	ImageOriginal         string                  `json:"image_original"`
	Title                 string                  `json:"title"`
	BotID                 string                  `json:"bot_id,omitempty"`
	ApiAppID              string                  `json:"api_app_id,omitempty"`
	StatusText            string                  `json:"status_text,omitempty"`
	StatusEmoji           string                  `json:"status_emoji,omitempty"`
	Team                  string                  `json:"team"`
	Fields                UserProfileCustomFields `json:"fields"`
}

// UserProfileCustomFields represents user profile's custom fields.
// Slack API's response data type is inconsistent so we use the struct.
// For detail, please see below.
// https://github.com/nlopes/slack/pull/298#discussion_r185159233
type UserProfileCustomFields struct {
	fields map[string]UserProfileCustomField
}

// UnmarshalJSON is the implementation of the json.Unmarshaler interface.
func (fields *UserProfileCustomFields) UnmarshalJSON(b []byte) error {
	// https://github.com/nlopes/slack/pull/298#discussion_r185159233
	if string(b) == "[]" {
		return nil
	}
	return json.Unmarshal(b, &fields.fields)
}

// MarshalJSON is the implementation of the json.Marshaler interface.
func (fields UserProfileCustomFields) MarshalJSON() ([]byte, error) {
	if len(fields.fields) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(fields.fields)
}

// ToMap returns a map of custom fields.
func (fields *UserProfileCustomFields) ToMap() map[string]UserProfileCustomField {
	return fields.fields
}

// Len returns the number of custom fields.
func (fields *UserProfileCustomFields) Len() int {
	return len(fields.fields)
}

// SetMap sets a map of custom fields.
func (fields *UserProfileCustomFields) SetMap(m map[string]UserProfileCustomField) {
	fields.fields = m
}

// FieldsMap returns a map of custom fields.
func (profile *UserProfile) FieldsMap() map[string]UserProfileCustomField {
	return profile.Fields.ToMap()
}

// SetFieldsMap sets a map of custom fields.
func (profile *UserProfile) SetFieldsMap(m map[string]UserProfileCustomField) {
	profile.Fields.SetMap(m)
}

// UserProfileCustomField represents a custom user profile field
type UserProfileCustomField struct {
	Value string `json:"value"`
	Alt   string `json:"alt"`
	Label string `json:"label"`
}

// User contains all the information of a user
type User struct {
	ID                string      `json:"id"`
	TeamID            string      `json:"team_id"`
	Name              string      `json:"name"`
	Deleted           bool        `json:"deleted"`
	Color             string      `json:"color"`
	RealName          string      `json:"real_name"`
	TZ                string      `json:"tz,omitempty"`
	TZLabel           string      `json:"tz_label"`
	Locale            string      `json:"locale"`
	TZOffset          int         `json:"tz_offset"`
	Profile           UserProfile `json:"profile"`
	IsBot             bool        `json:"is_bot"`
	IsAdmin           bool        `json:"is_admin"`
	IsOwner           bool        `json:"is_owner"`
	IsPrimaryOwner    bool        `json:"is_primary_owner"`
	IsRestricted      bool        `json:"is_restricted"`
	IsUltraRestricted bool        `json:"is_ultra_restricted"`
	IsStranger        bool        `json:"is_stranger"`
	IsAppUser         bool        `json:"is_app_user"`
	Has2FA            bool        `json:"has_2fa"`
	HasFiles          bool        `json:"has_files"`
	Presence          string      `json:"presence"`
}

// UserPresence contains details about a user online status
type UserPresence struct {
	Presence        string   `json:"presence,omitempty"`
	Online          bool     `json:"online,omitempty"`
	AutoAway        bool     `json:"auto_away,omitempty"`
	ManualAway      bool     `json:"manual_away,omitempty"`
	ConnectionCount int      `json:"connection_count,omitempty"`
	LastActivity    JSONTime `json:"last_activity,omitempty"`
}

type UserIdentityResponse struct {
	User UserIdentity `json:"user"`
	Team TeamIdentity `json:"team"`
	SlackResponse
}

type UserIdentity struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Image24  string `json:"image_24"`
	Image32  string `json:"image_32"`
	Image48  string `json:"image_48"`
	Image72  string `json:"image_72"`
	Image192 string `json:"image_192"`
	Image512 string `json:"image_512"`
}

type TeamIdentity struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	Image34       string `json:"image_34"`
	Image44       string `json:"image_44"`
	Image68       string `json:"image_68"`
	Image88       string `json:"image_88"`
	Image102      string `json:"image_102"`
	Image132      string `json:"image_132"`
	Image230      string `json:"image_230"`
	ImageDefault  bool   `json:"image_default"`
	ImageOriginal string `json:"image_original"`
}

type userResponseFull struct {
	Members      []User                  `json:"members,omitempty"` // ListUsers
	User         `json:"user,omitempty"` // GetUserInfo
	UserPresence                         // GetUserPresence
	SlackResponse
}

type UserSetPhotoParams struct {
	CropX int
	CropY int
	CropW int
}

func NewUserSetPhotoParams() UserSetPhotoParams {
	return UserSetPhotoParams{
		CropX: DEFAULT_USER_PHOTO_CROP_X,
		CropY: DEFAULT_USER_PHOTO_CROP_Y,
		CropW: DEFAULT_USER_PHOTO_CROP_W,
	}
}

func userRequest(ctx context.Context, path string, values url.Values, debug bool) (*userResponseFull, error) {
	response := &userResponseFull{}
	err := post(ctx, path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// GetUserPresence will retrieve the current presence status of given user.
func (api *Client) GetUserPresence(user string) (*UserPresence, error) {
	return api.GetUserPresenceContext(context.Background(), user)
}

// GetUserPresenceContext will retrieve the current presence status of given user with a custom context.
func (api *Client) GetUserPresenceContext(ctx context.Context, user string) (*UserPresence, error) {
	values := url.Values{
		"token": {api.Config.token},
		"user":  {user},
	}
	response, err := userRequest(ctx, "users.getPresence", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.UserPresence, nil
}

// GetUserInfo will retrieve the complete user information
func (api *Client) GetUserInfo(user string) (*User, error) {
	return api.GetUserInfoContext(context.Background(), user)
}

// GetUserInfoContext will retrieve the complete user information with a custom context
func (api *Client) GetUserInfoContext(ctx context.Context, user string) (*User, error) {
	values := url.Values{
		"token": {api.Config.token},
		"user":  {user},
	}
	response, err := userRequest(ctx, "users.info", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.User, nil
}

// GetUsers returns the list of users (with their detailed information)
func (api *Client) GetUsers() ([]User, error) {
	return api.GetUsersContext(context.Background())
}

// GetUsersContext returns the list of users (with their detailed information) with a custom context
func (api *Client) GetUsersContext(ctx context.Context) ([]User, error) {
	values := url.Values{
		"token":    {api.Config.token},
		"presence": {"1"},
	}
	response, err := userRequest(ctx, "users.list", values, api.debug)
	if err != nil {
		return nil, err
	}
	return response.Members, nil
}

// GetUserByEmail will retrieve the complete user information by email
func (api *Client) GetUserByEmail(email string) (*User, error) {
	return api.GetUserByEmailContext(context.Background(), email)
}

// GetUserByEmailContext will retrieve the complete user information by email with a custom context
func (api *Client) GetUserByEmailContext(ctx context.Context, email string) (*User, error) {
	values := url.Values{
		"token": {api.Config.token},
		"email": {email},
	}
	response, err := userRequest(ctx, "users.lookupByEmail", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.User, nil
}

// SetUserAsActive marks the currently authenticated user as active
func (api *Client) SetUserAsActive() error {
	return api.SetUserAsActiveContext(context.Background())
}

// SetUserAsActiveContext marks the currently authenticated user as active with a custom context
func (api *Client) SetUserAsActiveContext(ctx context.Context) error {
	values := url.Values{
		"token": {api.Config.token},
	}
	_, err := userRequest(ctx, "users.setActive", values, api.debug)

	return err
}

// SetUserPresence changes the currently authenticated user presence
func (api *Client) SetUserPresence(presence string) error {
	return api.SetUserPresenceContext(context.Background(), presence)
}

// SetUserPresenceContext changes the currently authenticated user presence with a custom context
func (api *Client) SetUserPresenceContext(ctx context.Context, presence string) error {
	values := url.Values{
		"token":    {api.Config.token},
		"presence": {presence},
	}
	_, err := userRequest(ctx, "users.setPresence", values, api.debug)

	return err
}

// GetUserIdentity will retrieve user info available per identity scopes
func (api *Client) GetUserIdentity() (*UserIdentityResponse, error) {
	return api.GetUserIdentityContext(context.Background())
}

// GetUserIdentityContext will retrieve user info available per identity scopes with a custom context
func (api *Client) GetUserIdentityContext(ctx context.Context) (*UserIdentityResponse, error) {
	values := url.Values{
		"token": {api.Config.token},
	}
	response := &UserIdentityResponse{}
	err := post(ctx, "users.identity", values, response, api.debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// SetUserPhoto changes the currently authenticated user's profile image
func (api *Client) SetUserPhoto(image string, params UserSetPhotoParams) error {
	return api.SetUserPhotoContext(context.Background(), image, params)
}

// SetUserPhotoContext changes the currently authenticated user's profile image using a custom context
func (api *Client) SetUserPhotoContext(ctx context.Context, image string, params UserSetPhotoParams) error {
	response := &SlackResponse{}
	values := url.Values{
		"token": {api.Config.token},
	}
	if params.CropX != DEFAULT_USER_PHOTO_CROP_X {
		values.Add("crop_x", strconv.Itoa(params.CropX))
	}
	if params.CropY != DEFAULT_USER_PHOTO_CROP_Y {
		values.Add("crop_y", strconv.Itoa(params.CropY))
	}
	if params.CropW != DEFAULT_USER_PHOTO_CROP_W {
		values.Add("crop_w", strconv.Itoa(params.CropW))
	}
	err := postLocalWithMultipartResponse(ctx, "users.setPhoto", image, "image", values, response, api.debug)
	if err != nil {
		return err
	}
	if !response.Ok {
		return errors.New(response.Error)
	}
	return nil
}

// DeleteUserPhoto deletes the current authenticated user's profile image
func (api *Client) DeleteUserPhoto() error {
	return api.DeleteUserPhotoContext(context.Background())
}

// DeleteUserPhotoContext deletes the current authenticated user's profile image with a custom context
func (api *Client) DeleteUserPhotoContext(ctx context.Context) error {
	response := &SlackResponse{}
	values := url.Values{
		"token": {api.Config.token},
	}
	err := post(ctx, "users.deletePhoto", values, response, api.debug)
	if err != nil {
		return err
	}
	if !response.Ok {
		return errors.New(response.Error)
	}
	return nil
}

// SetUserCustomStatus will set a custom status and emoji for the currently
// authenticated user. If statusEmoji is "" and statusText is not, the Slack API
// will automatically set it to ":speech_balloon:". Otherwise, if both are ""
// the Slack API will unset the custom status/emoji.
func (api *Client) SetUserCustomStatus(statusText, statusEmoji string) error {
	return api.SetUserCustomStatusContext(context.Background(), statusText, statusEmoji)
}

// SetUserCustomStatusContext will set a custom status and emoji for the currently authenticated user with a custom context
//
// For more information see SetUserCustomStatus
func (api *Client) SetUserCustomStatusContext(ctx context.Context, statusText, statusEmoji string) error {
	// XXX(theckman): this anonymous struct is for making requests to the Slack
	// API for setting and unsetting a User's Custom Status/Emoji. To change
	// these values we must provide a JSON document as the profile POST field.
	//
	// We use an anonymous struct over UserProfile because to unset the values
	// on the User's profile we cannot use the `json:"omitempty"` tag. This is
	// because an empty string ("") is what's used to unset the values. Check
	// out the API docs for more details:
	//
	// - https://api.slack.com/docs/presence-and-status#custom_status
	profile, err := json.Marshal(
		&struct {
			StatusText  string `json:"status_text"`
			StatusEmoji string `json:"status_emoji"`
		}{
			StatusText:  statusText,
			StatusEmoji: statusEmoji,
		},
	)

	if err != nil {
		return err
	}

	values := url.Values{
		"token":   {api.Config.token},
		"profile": {string(profile)},
	}

	response := &userResponseFull{}

	if err = post(ctx, "users.profile.set", values, response, api.debug); err != nil {
		return err
	}

	if !response.Ok {
		return errors.New(response.Error)
	}

	return nil
}

// UnsetUserCustomStatus removes the custom status message for the currently
// authenticated user. This is a convenience method that wraps (*Client).SetUserCustomStatus().
func (api *Client) UnsetUserCustomStatus() error {
	return api.UnsetUserCustomStatusContext(context.Background())
}

// UnsetUserCustomStatusContext removes the custom status message for the currently authenticated user
// with a custom context. This is a convenience method that wraps (*Client).SetUserCustomStatus().
func (api *Client) UnsetUserCustomStatusContext(ctx context.Context) error {
	return api.SetUserCustomStatusContext(ctx, "", "")
}

// GetUserProfile retrieves a user's profile information.
func (api *Client) GetUserProfile(userID string, includeLabels bool) (*UserProfile, error) {
	return api.GetUserProfileContext(context.Background(), userID, includeLabels)
}

type getUserProfileResponse struct {
	SlackResponse
	Profile *UserProfile `json:"profile"`
}

// GetUserProfileContext retrieves a user's profile information with a context.
func (api *Client) GetUserProfileContext(ctx context.Context, userID string, includeLabels bool) (*UserProfile, error) {
	values := url.Values{"token": {api.Config.token}, "user": {userID}}
	if includeLabels {
		values.Add("include_labels", "true")
	}
	resp := &getUserProfileResponse{}

	err := post(ctx, "users.profile.get", values, &resp, api.debug)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, errors.New(resp.Error)
	}
	return resp.Profile, nil
}