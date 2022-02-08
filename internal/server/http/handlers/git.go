package handlers

import (
	"bytes"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/git"
	"github.com/Jameslikestea/grm/internal/storage"
)

func Git(ctx *fiber.Ctx) error {
	ctx.Status(200)
	ctx.Write([]byte("200 OK\n"))
	return nil
}
func AdvertiseReference(stor storage.Storage) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		repo := fmt.Sprintf("%s.git", ctx.Params("*1"))
		service := ctx.Query("service")

		log.Info().Str("service", service).Str("repo", repo).Msg("Advertising references")

		switch service {
		case "git-upload-pack":
			ctx.Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
			ctx.Set("Cache-Control", "no-cache")
			refs, err := stor.ListReferences(repo)
			if err != nil {
				log.Warn().Err(err).Msg("Cannot list refs")
				ctx.Status(500)
				ctx.Write([]byte("Internal Server Error"))
			}
			ctx.Status(200)
			git.GenerateReferencePack(refs, true, service, ctx)
		default:
			ctx.Status(500)
			ctx.Write([]byte("Internal Server Error"))
		}
		return nil
	}
}

func UploadPack(stor storage.Storage) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		repo := fmt.Sprintf("%s.git", ctx.Params("*1"))

		log.Info().Str("service", "git-upload-pack").Msg("Uploading Pack")
		body := ctx.Body()
		wantReader := bytes.NewReader(body)
		haveReader := bytes.NewReader(body)

		w := git.WantList{}
		h := git.HaveList{}

		git.ParseWantList(w, wantReader)
		git.ParseHaveList(h, haveReader)

		wants := make([]plumbing.Hash, len(w))
		i := 0
		for k := range w {
			wants[i] = k
			i++
		}
		objs := git.GetNewObjects(stor, repo, wants, h)
		ctx.Set("Content-Type", "application/x-git-upload-pack-result")
		ctx.Set("Cache-Control", "no-cache")

		ctx.Write([]byte("0008NAK\n"))
		git.EncodePackfile(ctx, objs)

		return nil
	}
}
