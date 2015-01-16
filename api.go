package main

import (
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "github.com/martini-contrib/binding"
    "github.com/Lupino/collect/models"
    "github.com/go-xorm/xorm"
    "mime/multipart"
    "crypto/sha1"
    "io"
    "os"
    "encoding/hex"
    "net/http"
    "log"
)


type FileForm struct {
    File *multipart.FileHeader `form:"file"`
}

type TagForm struct {
    Tag  string                `form:"tag"`
}

type DatasetForm struct {
    FileForm
    TagForm
}

func api(mart *martini.ClassicMartini, engine *xorm.Engine) {
    mart.Post(API + "/dataset", binding.Bind(DatasetForm{}), func(form DatasetForm, r render.Render) {
        realFile, err := form.File.Open()
        if err != nil {
            log.Printf("Error: %s\n", err)
            r.JSON(http.StatusInternalServerError, map[string]interface{}{"err": "Upload file fail"})
            return
        }
        defer realFile.Close()
        hasher := sha1.New()
        io.Copy(hasher, realFile)
        fileKey := hex.EncodeToString(hasher.Sum(nil))
        var file = &models.File{Key: fileKey}
        has, _ := engine.Get(file)
        if !has {
            dst, err := os.Create(UPLOADPATH + fileKey)
            defer dst.Close()
            if err != nil {
                log.Printf("Error: %s\n", err)
                r.JSON(http.StatusInternalServerError, map[string]interface{}{"err": "Upload file fail"})
                return
            }
            realFile.Seek(0, 0)
            _, err = io.Copy(dst, realFile)
            if err != nil {
                log.Printf("Error: %s\n", err)
                r.JSON(http.StatusInternalServerError, map[string]interface{}{"err": "Upload file fail"})
                return
            }

            if _, err := engine.Insert(file); err != nil {
                log.Printf("Error: %s\n", err)
                r.JSON(http.StatusInternalServerError, map[string]interface{}{"err": "Upload file fail"})
                return
            }
        }

        var tag = &models.Tag{Name: form.Tag}
        has, _ = engine.Get(tag)
        if !has {
            if _, err := engine.Insert(tag); err != nil {
                log.Printf("Error: %s\n", err)
                r.JSON(http.StatusInternalServerError, map[string]interface{}{"err": "Save tag " + form.Tag + " fail"})
                return
            }
        }

        var dataset = &models.Dataset{FileId: file.Id, TagId: tag.Id}
        has, _ = engine.Get(dataset)
        if !has {
            if _, err := engine.Insert(dataset); err != nil {
                log.Printf("Error: %s\n", err)
                r.JSON(http.StatusInternalServerError, map[string]interface{}{
                    "err": "Save file tag: " + form.Tag + " error: " + err.Error()})
                return
            }
        }

        r.JSON(http.StatusOK, map[string]string{"msg": "ok"})
    })
}
