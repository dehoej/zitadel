package messages

import (
	"io"
	"mime/quotedprintable"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmail_GetContent_Body(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		wantContentType string
	}{
		{
			name:            "plain text",
			content:         "Hej Søren,\r\nbekræft venligst din e-mail.",
			wantContentType: "Content-Type: text/plain; charset=\"UTF-8\"\r\n",
		},
		{
			name:            "html",
			content:         "<html><body><p>Bekræft din e-mail: " + strings.Repeat("<span>æøå</span>", 100) + "</p></body></html>",
			wantContentType: "Content-Type: text/html; charset=\"UTF-8\"\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Email{
				Recipients:  []string{"to@example.com"},
				SenderEmail: "from@example.com",
				Subject:     "subject",
				Content:     tt.content,
			}
			got, err := msg.GetContent()
			require.NoError(t, err)

			header, body, ok := strings.Cut(got, "\r\n\r\n")
			require.True(t, ok)
			assert.Contains(t, header+"\r\n", tt.wantContentType)
			assert.Contains(t, header+"\r\n", "Content-Transfer-Encoding: quoted-printable\r\n")
			assert.NotContains(t, header, "Return-Path:")

			for _, line := range strings.Split(got, "\r\n") {
				assert.LessOrEqual(t, len(line), 76)
				for _, r := range line {
					assert.Less(t, r, rune(128))
				}
			}

			decoded, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(body)))
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(decoded))
		})
	}
}
