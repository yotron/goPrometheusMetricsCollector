/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package common

type Response struct {
	Status      int
	ContentType string
	ContentBody []byte
}

func (resp *Response) SetResponse(contentType string, contentBody []byte, status int) {
	resp.Status = status
	resp.ContentType = contentType
	resp.ContentBody = contentBody
}

func (resp *Response) SetErrorResponse() {
	resp.Status = 500
	resp.ContentType = "text/plain"
	resp.ContentBody = []byte("alles Mist")
}
