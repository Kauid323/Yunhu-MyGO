package app

import (
	"bytes"
	"image"
	// The formats a picture may come in: the header is all that is read of
	// them here, to know how big the message will be before it is sent.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// imageExts are the files the picture picker offers. A picture is sent as a
// picture and everything else as a file, and the extension is what says
// which: a file picked through the other button is still sent as what it is,
// so the two buttons differ in what they show rather than in what they do.
var imageExts = []string{"png", "jpg", "jpeg", "gif", "webp", "bmp"}

// videoExts are the clips the video picker offers. A clip has a bucket of
// its own, so an mp4 sent through the file button would land where no
// player looks; the name is what says which bucket a file belongs in.
var videoExts = []string{"mp4", "avi", "mov", "mkv", "flv", "wmv", "3gp"}

// pickImages asks for pictures, and sends the ones that were picked.
func (a *App) pickImages() {
	a.pick(mygo.FileFilter{Name: i18n.T("attach.images"), Extensions: imageExts})
}

// pickVideos asks for clips, and sends the ones that were picked.
func (a *App) pickVideos() {
	a.pick(mygo.FileFilter{Name: i18n.T("attach.videos"), Extensions: videoExts})
}

// pickFiles asks for any file, and sends the ones that were picked.
func (a *App) pickFiles() {
	a.pick(mygo.FileFilter{Name: i18n.T("attach.files"), Extensions: []string{"*"}})
}

// pick opens the desktop's own file dialog. It is asked for off the drawing
// thread: the dialog waits for a person, and a frame that waits for one is a
// window that stops drawing until it is answered.
func (a *App) pick(filter mygo.FileFilter) {
	chatID, chatType := a.current, a.currentType
	if chatID == "" {
		return
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:    i18n.T("attach.pick"),
			Filters:  []mygo.FileFilter{filter},
			Multiple: true,
		})
		if err != nil {
			a.update(func() { a.note = errorLine(err) })
			return
		}
		a.update(func() {
			// The chat that was open when the dialog was asked for is the
			// chat the files were picked for: picking one while the dialog
			// was up does not send them somewhere else.
			if a.current != chatID {
				return
			}
			for _, path := range paths {
				a.attach(chatID, chatType, path)
			}
		})
	}()
}

// attach sends one file: the message is put in the history the moment the
// file is picked, and the upload fills it in. A message that is on its way
// shows as one, and one that could not be sent says why, as a text does.
func (a *App) attach(chatID string, chatType int, path string) {
	name := filepath.Base(path)
	video := isVideoName(name)
	picture := !video && isImageName(name)
	size := int64(0)
	if info, err := os.Stat(path); err == nil {
		size = info.Size()
	}
	pending := &yh.Message{
		MsgID:       "local-" + yh.NewID(),
		Sender:      yh.Sender{ChatID: a.me.ID, Name: a.me.Name, AvatarURL: a.me.AvatarURL, ChatType: yh.ChatUser},
		ContentType: yh.ContentFile,
		TimestampMS: uint64(time.Now().UnixMilli()),
		Pending:     true,
	}
	switch {
	case video:
		pending.ContentType = yh.ContentVideo
		pending.Content.FileName, pending.Content.FileSize = name, size
	case picture:
		pending.ContentType = yh.ContentImage
	default:
		pending.Content.FileName, pending.Content.FileSize = name, size
	}
	a.msgs = append(a.msgs, pending)
	a.trimMessages()
	a.jumpToEnd = true

	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			a.failAttach(pending.MsgID, err)
			return
		}
		var sent *yh.Message
		switch {
		case video:
			key, err := a.client.UploadVideo(name, data)
			if err == nil {
				sent, err = a.client.SendVideo(chatID, chatType, key, name, int64(len(data)))
			}
			if err != nil {
				a.failAttach(pending.MsgID, err)
				return
			}
		case picture:
			width, height := imageSize(data)
			key, err := a.client.UploadImage(name, data)
			if err == nil {
				sent, err = a.client.SendImage(chatID, chatType, key, int64(width), int64(height), int64(len(data)))
			}
			if err != nil {
				a.failAttach(pending.MsgID, err)
				return
			}
		default:
			key, err := a.client.UploadFile(name, data)
			if err == nil {
				sent, err = a.client.SendFile(chatID, chatType, key, name, int64(len(data)))
			}
			if err != nil {
				a.failAttach(pending.MsgID, err)
				return
			}
		}
		preview := i18n.T("msg.fileBare")
		switch {
		case video:
			preview = i18n.T("msg.video")
		case picture:
			preview = i18n.T("msg.image")
		}
		a.update(func() {
			for i, m := range a.msgs {
				if m.MsgID != pending.MsgID {
					continue
				}
				sent.Sender = pending.Sender
				a.msgs[i] = sent
				break
			}
			a.touchConv(chatID, preview)
		})
	}()
}

// failAttach marks a message that could not be sent, and says why.
func (a *App) failAttach(id string, err error) {
	a.update(func() {
		for _, m := range a.msgs {
			if m.MsgID != id {
				continue
			}
			m.Pending = false
			m.Failed = err.Error()
			break
		}
		a.note = errorLine(err)
	})
}

// isImageName reports whether a file is a picture: what the message shows
// depends on it, and the name is all that is known before the file is read.
func isImageName(name string) bool {
	return hasExt(name, imageExts)
}

// isVideoName reports whether a file is a clip. A clip goes to the video
// bucket rather than the file one, so the name has to say so before the
// bytes are read.
func isVideoName(name string) bool {
	return hasExt(name, videoExts)
}

func hasExt(name string, exts []string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	for _, e := range exts {
		if e == ext {
			return true
		}
	}
	return false
}

// imageSize is how big a picture is, read from its header rather than from
// the whole of it: a photograph of several megabytes says how big it is in
// its first few hundred bytes, and a picture that cannot be read is sent
// without a size rather than not sent at all.
func imageSize(data []byte) (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// pendingPicture is a picture that is still being sent: the room it will
// take, with the word for waiting in it, as a chat shows one.
func (a *App) pendingPicture(c *ui.Context, m *yh.Message) {
	a.pendingMedia(c, m, "msg.imageLoading")
}

// pendingVideo is a clip that is still being sent, as a picture is.
func (a *App) pendingVideo(c *ui.Context, m *yh.Message) {
	a.pendingMedia(c, m, "msg.videoLoading")
}

// pendingMedia is the room a picture or a clip will take while it is on its
// way, with the word for waiting in it.
func (a *App) pendingMedia(c *ui.Context, m *yh.Message, wait string) {
	ui.Column(c).Width(pictureBoxDIP * 0.62).Height(pictureBoxDIP * 0.44).
		Radius(6).Background(qq.inputBg).Justify(ui.Center).AlignItems(ui.Center).
		Children(func() {
			if m.ContentType == yh.ContentVideo {
				ui.Icon(c, iconVideo).FontSize(20).TextColor(qq.textFaint)
			}
			ui.Text(c, i18n.T(wait)).FontSize(11).TextColor(qq.textFaint)
		})
	a.bubbleState(c, m, qq.text, 0)
}
