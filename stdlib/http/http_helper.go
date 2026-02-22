// Hand-written Go helpers for stdlib/http.
// Functions here require Go closures or type assertions not yet expressible
// in Kukicha syntax.  They must NOT duplicate any function in http.go.

package http

import "net/http"

// SecureHeaders returns middleware that injects security response headers before
// delegating to the wrapped handler.  Use with http.Serve:
//
//	http.Serve(":8080", httphelper.SecureHeaders(mux))
func SecureHeaders(handler any) any {
	h := handler.(http.Handler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		h.ServeHTTP(w, r)
	})
}
