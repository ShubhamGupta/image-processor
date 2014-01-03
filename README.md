image-processor
===============

A processor written in Go, receives images(Image Path Actually) from RabbitMQ exchange and crops the images based on the passed in parameters.

This is a script written in Go-Lang that reads in messages from a rabbitMQ Exchange. The messages are JSON objects Like:

{
  path: <Path-To_Image>
  crop_x: <x-origin>
  crop_y: <y-origin>
  crop_w: <Image-Width>
  crop_h: <Image-Height>
}

Steps:

1) It receives the message from the queue and passes it in a GoRoutine for cropping the image.

2) The GoRoutine uses imagemagick image processing tool to crop the image. We use the 'magick' library to hook into the imagemagick API.

3) The GoRoutine returms after cropping the image.

There is a companion Rails App specifically for this script located at https://github.com/aditya-kapoor/image-cropping.
This application simply allows you to upload images and crop them. To test the app and the script:

1) Run the rails app on your local environment.

2) Run the image-processor Go script. To do this, you must have RabbitMQ installed on your system and running.
This will start listening to the exchange and intercept the messages on it.

3) Open your Rails app(http://localhost:3000), upload an image and start cropping.

*Note*

* This app is intended only for testing purposes. It saves the cropped image in public directory. So, we have to go with the default paperclip config, i.e.: we cant store the attachments on, say, S3.

* The Image is stored in '200x200+100+100' folder(i.e.: Based on the dimensions of the cropped image: WxH+X+Y) in paperclip save directory.
