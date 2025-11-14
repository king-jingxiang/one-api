package render

import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/songquanpeng/one-api/common"
    "github.com/songquanpeng/one-api/common/logger"
)

func StringData(c *gin.Context, str string) {
    str = strings.TrimPrefix(str, "data: ")
    str = strings.TrimSuffix(str, "\r")

    if str == "[DONE]" {
        buf := c.GetString("stream_log_buf")
        if buf != "" {
            logger.SysLogf("StreamData: %s", buf)
            c.Set("stream_log_buf", "")
        }
    } else {
        var contentPiece string
        var obj map[string]any
        if err := json.Unmarshal([]byte(str), &obj); err == nil {
            if choices, ok := obj["choices"].([]any); ok {
                for _, ch := range choices {
                    if m, ok := ch.(map[string]any); ok {
                        if delta, ok := m["delta"].(map[string]any); ok {
                            if cont, ok := delta["content"]; ok {
                                switch v := cont.(type) {
                                case string:
                                    contentPiece += v
                                case []any:
                                    for _, part := range v {
                                        if pm, ok := part.(map[string]any); ok {
                                            if t, ok := pm["text"].(string); ok {
                                                contentPiece += t
                                            }
                                        }
                                    }
                                }
                            }
                        } else if text, ok := m["text"].(string); ok {
                            contentPiece += text
                        }
                    }
                }
            }
        } else {
            contentPiece = str
        }
        if len(contentPiece) > 0 {
            c.Set("stream_log_buf", c.GetString("stream_log_buf")+contentPiece)
        }
    }

    c.Render(-1, common.CustomEvent{Data: "data: " + str})
    c.Writer.Flush()
}

func ObjectData(c *gin.Context, object interface{}) error {
	jsonData, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("error marshalling object: %w", err)
	}
	StringData(c, string(jsonData))
	return nil
}

func Done(c *gin.Context) {
	StringData(c, "[DONE]")
}
