package service

import (
	"bufio"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HelperService struct {
}

func (s *HelperService) ChunkData(ctx *fiber.Ctx, data [][]byte) error {

	ctx.Set("Transfer-Encoding", "chunked")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {

		for i := 0; i < len(data); i++ {
			_, err := w.Write(data[i])
			if err != nil {
				return
			}
			err = w.Flush()
			if err != nil {
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
	return nil
}
