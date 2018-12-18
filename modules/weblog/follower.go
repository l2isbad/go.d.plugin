package weblog

import (
	"io"

	"github.com/hpcloud/tail"
)

type follower interface {
	lines() chan *tail.Line
	stop()
}

func newFollower(filename string) (follower, error) {
	f := &followerImp{filename: filename}

	if err := f.start(); err != nil {
		return nil, err
	}

	return f, nil
}

type followerImp struct {
	filename string
	tail     *tail.Tail
}

func (f *followerImp) lines() chan *tail.Line {
	return f.tail.Lines
}

func (f *followerImp) stop() {
	f.tail.Cleanup()
	_ = f.tail.Stop()

}

func (f *followerImp) start() (err error) {
	f.tail, err = tail.TailFile(
		f.filename,
		tail.Config{
			Follow:    true,
			ReOpen:    true,
			MustExist: true,
			Location:  &tail.SeekInfo{Whence: io.SeekEnd},
			Logger:    tail.DiscardingLogger,
		})

	return err
}
