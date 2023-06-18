package server

import (
	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	// types "github.com/0187773933/StreamDeck/v1/types"
)

// var ui_html_pages = map[ string ]string {
// 	"/": "./v1/server/html/admin.html" ,
// 	"/users": "./v1/server/html/admin_view_users.html" ,
// 	// "/user/new": "./v1/server/html/admin_user_new.html" ,
// 	"/user/new/handoff/:uuid": "./v1/server/html/admin_user_new_handoff.html" ,
// 	"/user/checkin": "./v1/server/html/admin_user_checkin.html" ,
// 	"/user/checkin/:uuid": "./v1/server/html/admin_user_checkin.html" ,
// 	"/user/checkin/:uuid/edit": "./v1/server/html/admin_user_checkin.html" ,
// 	"/user/checkin/new": "./v1/server/html/admin_user_checkin.html" ,
// 	"/user/edit/:uuid": "./v1/server/html/admin_user_edit.html" ,
// 	"/checkins": "./v1/server/html/admin_view_total_checkins.html" ,
// 	"/emails": "./v1/server/html/admin_view_all_emails.html" ,
// 	"/phone-numbers": "./v1/server/html/admin_view_all_phone_numbers.html" ,
// 	"/barcodes": "./v1/server/html/admin_view_all_barcodes.html" ,
// 	"/sms": "./v1/server/html/admin_sms_all_users.html" ,
// 	"/email": "./v1/server/html/admin_email_all_users.html" ,
// }

func ( s *Server ) PressButton( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	// // fmt.Println( context.GetReqHeaders() )
	// email_message := context.FormValue( "email_message" )

	// db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()
	// db.View( func( tx *bolt_api.Tx ) error {
	// 	bucket := tx.Bucket( []byte( "users" ) )
	// 	bucket.ForEach( func( uuid , value []byte ) error {
	// 		var viewed_user user.User
	// 		decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
	// 		json.Unmarshal( decrypted_bucket_value , &viewed_user )
	// 		if viewed_user.EmailAddress == "" { return nil; }
	// 		fmt.Printf( "%s === %s\n" , "from@example.com" , viewed_user.EmailAddress )
	// 		return nil
	// 	})
	// 	return nil
	// })

	s.UI.ActivePageID = "spotify-triple"
	s.UI.Render()

	return context.JSON( fiber.Map{
		"route": "/some-button-number" ,
		"result": "temp" ,
	})
}

func ( s *Server ) SetupRoutes() {

	// admin_route_group := s.FiberApp.Group( "/admin" )

	// // HTML UI Pages
	// admin_route_group.Get( "/login" , ServeLoginPage )
	// for url , _ := range ui_html_pages {
	// 	admin_route_group.Get( url , ServeAuthenticatedPage )
	// }

	button_max := 100
	for i := 0; i < button_max; i++ {
		s.FiberApp.Get( fmt.Sprintf( "/%d" , ( i + 1 ) ) , s.PressButton )
	}

}