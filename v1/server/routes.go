package server

import (
	"fmt"
	"strings"
	// "time"
	filepath "path/filepath"
	fiber "github.com/gofiber/fiber/v2"
	strconv "strconv"
	// types "github.com/0187773933/StreamDeck/v1/types"
	bolt_api "github.com/boltdb/bolt"
	ui_wrapper "github.com/0187773933/StreamDeck/v1/ui"
)

func ( s *Server ) PressButton( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }

	// s.UI.ActivePageID = "spotify-triple"
	// s.UI.Render()

	button_type := ""
	button_int , button_int_err := strconv.Atoi( context.Params( "button" ) )
	if button_int_err == nil {
		// Check if the integer is within the uint8 range (0-255)
		if button_int >= 0 && button_int <= 255 {
			button_type = "number"
		}
	} else {
		button_type = "string"
	}

	switch button_type {
		case "number":
			s.UI.SingleClickNumber( uint8( button_int ) )
			break;
		case "string":
			s.UI.SingleClickId( context.Params( "button" ) )
			break;
	}

	return context.JSON( fiber.Map{
		"route": "/:button" ,
		"type": button_type ,
		"id": context.Params( "button" ) ,
		"result": "success" ,
	})
}

func ( s *Server ) DecreaseBrightness( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.DecreaseBrightness()
	return context.JSON( fiber.Map{
		"route": "/brightness/decrease" ,
		"brightness": s.UI.Brightness ,
		"result": "success" ,
	})
}

func ( s *Server ) IncreaseBrightness( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.IncreaseBrightness()
	return context.JSON( fiber.Map{
		"route": "/brightness/increase" ,
		"brightness": s.UI.Brightness ,
		"result": "success" ,
	})
}

func ( s *Server ) Show( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.Show()
	return context.JSON( fiber.Map{
		"route": "/show" ,
		"result": "success" ,
	})
}

func ( s *Server ) Hide( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.Hide()
	return context.JSON( fiber.Map{
		"route": "/hide" ,
		"result": "success" ,
	})
}

func ( s *Server ) Mute( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.Mute()
	return context.JSON( fiber.Map{
		"route": "/mute" ,
		"result": "success" ,
		"muted": s.UI.Muted ,
	})
}

func ( s *Server ) UnMute( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.UnMute()
	return context.JSON( fiber.Map{
		"route": "/unmute" ,
		"result": "success" ,
		"muted": s.UI.Muted ,
	})
}

func ( s *Server ) Sleep( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.Hide()
	s.UI.Mute()
	return context.JSON( fiber.Map{
		"route": "/sleep" ,
		"result": "success" ,
		"muted": s.UI.Muted ,
	})
}

func ( s *Server ) Wake( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	s.UI.Show()
	s.UI.UnMute()
	return context.JSON( fiber.Map{
		"route": "/wake" ,
		"result": "success" ,
		"muted": s.UI.Muted ,
	})
}

func ( s *Server ) RenderPage( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	page_id := context.Params( "id" )
	s.UI.Clear()

	// s.UI.SetActivePageID( string( page_id ) )
	// race condition ????
	s.UI.DB.Update( func( tx *bolt_api.Tx ) error {
		tmp2_bucket , _ := tx.CreateBucketIfNotExists( []byte( "tmp2" ) )
		tmp2_bucket.Put( []byte( "active-page-id" ) , []byte( page_id ) )
		return nil
	})
	s.UI.Render()
	return context.JSON( fiber.Map{
		"route": "/page/:id" ,
		"id": page_id ,
		"result": "success" ,
	})
}

func ( s *Server ) GetPageAddTiledImage( context *fiber.Ctx ) ( error ) {
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Upload Image</title>
	</head>
	<body>
		<form action="/page/add/tiled" method="post" enctype="multipart/form-data">
			<input type="file" name="image" required /><br><br>
			<input type="text" name="MP3String" placeholder="MP3 String" /><br><br>
			<input type="text" name="SingleClickString" placeholder="Single Click Command" /><br><br>
			<input type="text" name="ReturnPageString" placeholder="Return Page" /><br><br>
			<input type="submit" value="Upload" />
		</form>
	</body>
	</html>
	`
	context.Set( "Content-Type" , "text/html" )
	return context.SendString(htmlContent)
}

func ( s *Server ) PageAddTiledImage( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }

	// Parse the multipart form:
	form, err := context.MultipartForm()
	if err != nil {
		return err
	}

	// Get the first file from the "image" key:
	files := form.File["image"]
	if len(files) == 0 {
		return fiber.ErrBadRequest
	}

	file := files[0]

	// Save the file to the "images" directory:
	filename := filepath.Base(file.Filename)
	targetPath := filepath.Join("./images", filename)
	if err := context.SaveFile(file, targetPath); err != nil {
		return err
	}

	// Process the image:
	button := ui_wrapper.Button{
		MP3:         context.FormValue("MP3String"),
		SingleClick: context.FormValue("SingleClickString"),
		ReturnPage:  context.FormValue("ReturnPageString"),
	}

	page_id := s.UI.AddImageAsTiledButton(targetPath, button)

	return context.JSON( fiber.Map{
		"route": "/page/add/tiled" ,
		"page_id": page_id ,
		"result": "success" ,
	})
}


func ( s *Server ) GetButtonAdd( context *fiber.Ctx ) ( error ) {
	html_content := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Add Button</title>
		<script type="text/javascript">
			document.addEventListener("DOMContentLoaded", function() {
				var counter = 1;
				document.getElementById("addOption").addEventListener("click", function() {
					var container = document.getElementById("optionsContainer");
					var inputKey = document.createElement("input");
					inputKey.type = "text";
					inputKey.name = "OptionKey-" + counter;
					inputKey.placeholder = "Option Key";
					var inputValue = document.createElement("input");
					inputValue.type = "text";
					inputValue.name = "OptionValue-" + counter;
					inputValue.placeholder = "Option Value";
					container.appendChild(inputKey);
					container.appendChild(inputValue);
					container.appendChild(document.createElement("br"));
					counter++;
				});
			});
		</script>
	</head>
	<body>
		<form action="/button/add" method="post" enctype="multipart/form-data">
			<input type="file" name="image" /><br><br>
			<input type="text" name="ID" placeholder="Button-ID" /><br><br>
			<input type="text" name="MP3String" placeholder="MP3 String" /><br><br>
			<input type="text" name="SingleClickString" placeholder="Single Click Command" /><br><br>
			<input type="text" name="DoubleClickString" placeholder="Double Click Command" /><br><br>
			<input type="text" name="TripleClickString" placeholder="Triple Click Command" /><br><br>
			<input type="text" name="ReturnPageString" placeholder="Return Page" /><br><br>
			<div id="optionsContainer"></div>
			<button type="button" id="addOption">Add Option</button><br><br>
			<input type="submit" value="Upload" />
		</form>
	</body>
	</html>
	`
	context.Set( "Content-Type" , "text/html" )
	return context.SendString( html_content )
}

func (s *Server) GetButtonEdit(context *fiber.Ctx) error {
    button_id := context.Params( "id" )
    fmt.Println( "trying to edit" , button_id )
    _ , exists := s.UI.Buttons[ button_id ]
    if exists == false {
        return context.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Button not found"})
    }
    button := s.UI.Buttons[ button_id ]
    // Generate the HTML form with button data filled in
    htmlContent := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>Edit Button</title>
    </head>
    <body>
        <form action="/button/%s/edit" method="post" enctype="multipart/form-data">
            <input type="file" name="image" /><br><br> <!-- Allow image update, no 'required' if image already exists -->
            <input type="text" name="ID" value="%s" readonly /><br><br> <!-- ID is readonly to prevent changes -->
            <input type="text" name="MP3String" value="%s" placeholder="MP3 String" /><br><br>
            <input type="text" name="SingleClickString" value="%s" placeholder="Single Click Command" /><br><br>
            <input type="text" name="DoubleClickString" value="%s" placeholder="Double Click Command" /><br><br>
            <input type="text" name="TripleClickString" value="%s" placeholder="Triple Click Command" /><br><br>
            <input type="text" name="ReturnPageString" value="%s" placeholder="Return Page" /><br><br>
            <input type="submit" value="Update" />
        </form>
    </body>
    </html>
    `, button_id, button_id, button.MP3, button.SingleClick, button.DoubleClick, button.TripleClick, button.ReturnPage)

    context.Set("Content-Type", "text/html")
    return context.SendString(htmlContent)
}

func ( s *Server ) ButtonAdd( context *fiber.Ctx ) ( error ) {
	if validate_admin( context ) == false { return serve_failed_attempt( context ) }
	form , err := context.MultipartForm()
	if err != nil { return err }
	files := form.File[ "image" ]
	var image_path string
	if len( files ) == 0 {
		// use default icon image if not sent
		image_path = filepath.Join( "./images" , "1f49a.png" )
	} else {
		file := files[ 0 ]
		filename := filepath.Base( file.Filename )
		image_path = filepath.Join( "./images" , filename )
		err = context.SaveFile( file , image_path );
		if err != nil { return err }
	}

	button_id := context.FormValue( "ID" )
    options := make(map[string]string)
    for key, values := range form.Value {
        if strings.HasPrefix( key , "OptionKey-" ) {
            optionIndex := strings.TrimPrefix( key , "OptionKey-" )
            valueKey := "OptionValue-" + optionIndex
            if value, ok := form.Value[valueKey]; ok && len(value) > 0 {
                options[values[0]] = value[0]
            }
        }
    }

	button := ui_wrapper.Button{
		Id: button_id ,
		MP3: context.FormValue( "MP3String" ) ,
		Image: image_path ,
		SingleClick: context.FormValue( "SingleClickString" ) ,
		DoubleClick: context.FormValue( "DoubleClickString" ) ,
		TripleClick: context.FormValue( "TripleClickString" ) ,
		ReturnPage: context.FormValue( "ReturnPageString" ) ,
		Options: options ,
	}

	s.UI.AddButton( button_id , button )

	return context.JSON( fiber.Map{
		"route": "/button/add" ,
		"result": "success" ,
		"id": button_id ,
		"button": button ,
	})
}

func ( s *Server ) SetupRoutes() {

	// admin_route_group := s.FiberApp.Group( "/admin" )

	// // HTML UI Pages
	// admin_route_group.Get( "/login" , ServeLoginPage )
	// for url , _ := range ui_html_pages {
	// 	admin_route_group.Get( url , ServeAuthenticatedPage )
	// }

	s.FiberApp.Get( "/" , func( ctx *fiber.Ctx ) ( error ) {
		ctx.Set( "Content-Type" , "text/html" )
		return ctx.SendString( "<h1>Stream Deck Server</h1>" )
	})

	s.FiberApp.Get( "/show" , s.Show )
	s.FiberApp.Get( "/brightness/increase" , s.IncreaseBrightness )
	s.FiberApp.Get( "/brightness/decrease" , s.DecreaseBrightness )
	s.FiberApp.Get( "/hide" , s.Hide )
	s.FiberApp.Get( "/mute" , s.Mute )
	s.FiberApp.Get( "/unmute" , s.UnMute )
	s.FiberApp.Get( "/sleep" , s.Sleep )
	s.FiberApp.Get( "/wake" , s.Wake )
	s.FiberApp.Get( "/:button" , s.PressButton )
	s.FiberApp.Get( "/button/:id/edit" , s.GetButtonEdit )
	s.FiberApp.Get( "/button/add" , s.GetButtonAdd )
	s.FiberApp.Post( "/button/add" , s.ButtonAdd )
	s.FiberApp.Get( "/page/:id" , s.RenderPage )
	s.FiberApp.Get( "/page/add/tiled" , s.GetPageAddTiledImage )
	s.FiberApp.Post( "/page/add/tiled" , s.PageAddTiledImage )

}