package ghproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func ghProxyHandler(c *gin.Context, withSecurity bool) {
	remote, err := url.Parse("https://api.github.com")
	if err != nil {
		return
	}
	//note your github personal key should be in the GITHUB_ACCESS_TOKEN environment
	//variable, im using a helper, see main(), and a .env file
	authHeader := "Token " + os.Getenv("GITHUB_ACCESS_TOKEN")

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host

		if withSecurity {
			req.Header.Set("Authorization", authHeader)
		}
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("ghapi")
	}

	//This is a slight little hack because we are using the Go Gin Library.  This library
	//adds CORS headers, but so does GitHub.  We need to remove the header here and allow
	//Git to add them back or there will be two redundant Access-Control-Allow-Origin headers
	//which is not allowed and the browser will complain
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Access-Control-Allow-Origin")
		return nil
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func GhProxy(c *gin.Context) {
	ghProxyHandler(c, false)
}

func GhSecureProxy(c *gin.Context) {
	ghProxyHandler(c, true)
}
