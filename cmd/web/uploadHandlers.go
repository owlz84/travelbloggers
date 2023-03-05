package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"strings"
)

type ImageDetail struct {
	ImagePath     string
	ThumbnailPath string
}

type imageUploadForm struct {
	Images []*ImageDetail
}

func (app *application) uploadImage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/uploadImages.tmpl",
	)

}

func (app *application) uploadImagePost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(1024 << 20)
	if err != nil {
		app.serverError(w, err)
		return
	} // 32MB is the default used by FormFile
	fhs := r.MultipartForm.File["img"]
	var form imageUploadForm
	for _, fh := range fhs {
		f, err := fh.Open()
		if err != nil {
			app.serverError(w, err)
			return
		}
		imageRoot := "./ui/static/user-img"
		tempFile, err := os.CreateTemp(imageRoot, "upload-*.png")
		if err != nil {
			app.serverError(w, err)
			return
		}
		fileName := tempFile.Name()

		fileBytes, err := io.ReadAll(f)
		if err != nil {
			fmt.Println(err)
		}
		tempFile.Write(fileBytes)
		tempFile.Close()

		var thumbnail image.Image
		img, err := imaging.Open(fileName)
		if err != nil {
			app.serverError(w, err)
			return
		}

		thumbnail = imaging.Thumbnail(img, 256, 256, imaging.CatmullRom)
		dst := imaging.New(256, 256, color.NRGBA{0, 0, 0, 0})
		dst = imaging.Paste(dst, thumbnail, image.Pt(0, 0))
		thumbnailPath := strings.ReplaceAll(fileName, "user-img", "user-img/thumbs")
		err = imaging.Save(dst, thumbnailPath)
		if err != nil {
			app.serverError(w, err)
			return
		}

		image := ImageDetail{
			ImagePath:     strings.ReplaceAll(fileName, "./ui", ""),
			ThumbnailPath: strings.ReplaceAll(thumbnailPath, "./ui", "")}
		form.Images = append(form.Images, &image)

		f.Close()
	}

	data := app.newTemplateData(r)
	data.Form = form
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/uploadImages.tmpl",
	)

}
