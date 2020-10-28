package sections

import (
	"testing"

	"github.com/jonaslu/ain/data"
)

func TestParseHostHappyPath(t *testing.T) {
	parsedTepmlate := &data.TemplateSections{}

	parseSection(`
[Variables]
ENV=test
GOAT=${TEST}
UUIDS=kjsdf
[Host]
http://localhost:8080/
http://very.important.url.se/
[URL]
[Headers]
-H $TOKEN
[Method]
GET
[QPs]
[Body]
[Backend]
curl
-sS
--no-buffer`, parsedTepmlate)

}
