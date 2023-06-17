package adminroutes

import (
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/StreamDeck/v1/types"
)

var GlobalConfig *types.ConfigFile

var ui_html_pages = map[ string ]string {
	"/": "./v1/server/html/admin.html" ,
	"/users": "./v1/server/html/admin_view_users.html" ,
	// "/user/new": "./v1/server/html/admin_user_new.html" ,
	"/user/new/handoff/:uuid": "./v1/server/html/admin_user_new_handoff.html" ,
	"/user/checkin": "./v1/server/html/admin_user_checkin.html" ,
	"/user/checkin/:uuid": "./v1/server/html/admin_user_checkin.html" ,
	"/user/checkin/:uuid/edit": "./v1/server/html/admin_user_checkin.html" ,
	"/user/checkin/new": "./v1/server/html/admin_user_checkin.html" ,
	"/user/edit/:uuid": "./v1/server/html/admin_user_edit.html" ,
	"/checkins": "./v1/server/html/admin_view_total_checkins.html" ,
	"/emails": "./v1/server/html/admin_view_all_emails.html" ,
	"/phone-numbers": "./v1/server/html/admin_view_all_phone_numbers.html" ,
	"/barcodes": "./v1/server/html/admin_view_all_barcodes.html" ,
	"/sms": "./v1/server/html/admin_sms_all_users.html" ,
	"/email": "./v1/server/html/admin_email_all_users.html" ,
}

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	admin_route_group := fiber_app.Group( "/admin" )

	// HTML UI Pages
	admin_route_group.Get( "/login" , ServeLoginPage )
	for url , _ := range ui_html_pages {
		admin_route_group.Get( url , ServeAuthenticatedPage )
	}

	// API Routes
	admin_route_group.Get( "/logout" , Logout )
	admin_route_group.Post( "/login" , HandleLogin )

	admin_route_group.Post( "/user/new" , HandleNewUserJoin )
	admin_route_group.Post( "/user/edit" , HandleUserEdit )
	admin_route_group.Get( "/user/delete/:uuid" , DeleteUser )
	// admin_route_group.Get( "/user/check/username" , CheckIfFirstNameLastNameAlreadyExists )

	admin_route_group.Post( "/user/checkin/:uuid" , UserCheckIn )
	admin_route_group.Get( "/user/checkin/test/:uuid" , UserCheckInTest )

	admin_route_group.Get( "/user/get/all" , GetAllUsers )
	admin_route_group.Get( "/user/get/all/checkins" , GetAllCheckIns )
	admin_route_group.Get( "/user/get/all/emails" , GetAllEmails )
	admin_route_group.Get( "/user/get/all/phone-numbers" , GetAllPhoneNumbers )
	admin_route_group.Get( "/user/get/all/barcodes" , GetAllBarcodes )
	admin_route_group.Get( "/user/get/:uuid" , GetUser )
	admin_route_group.Get( "/user/get/barcode/:barcode" , GetUserViaBarcode )

	admin_route_group.Get( "/user/search/username/:username" , UserSearch )
	admin_route_group.Get( "/user/search/username/fuzzy/:username" , UserSearchFuzzy )
	admin_route_group.Get( "/print-test" , PrintTest )

	admin_route_group.Get( "/user/reports/main" , GetReportMain )
	admin_route_group.Get( "/user/reports/mail-chimp" , GetReportMailChimp )

	admin_route_group.Post( "/user/sms/all" , SMSAllUsers )
	admin_route_group.Post( "/user/email/all" , EmailAllUsers )

}