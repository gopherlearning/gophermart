package postgres

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadSeeds(t *testing.T) {

	Convey("Загружаем сиды", t, func() {
		dburl, respID, err := startDB("test-gophermart", "postgres:14")
		So(err, ShouldBeNil)
		So(respID, ShouldNotBeBlank)
		So(dburl, ShouldNotBeBlank)
		store, err := NewStorage(dburl, logrus.StandardLogger(), "secret")
		So(err, ShouldBeNil)
		So(store, ShouldNotBeNil)
		time.Sleep(10 * time.Second)
		err = stopDB(respID)
		So(err, ShouldBeNil)
	})
}
