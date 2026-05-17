package middleware

import (
	"net"
	"net/http"
)

type trustedContextKey string

const TrustedSubnetKey trustedContextKey = "trusted_subnet"

func TrustedSubnet(subnetCIDR string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnetCIDR == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, trustedNet, err := net.ParseCIDR(subnetCIDR)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			clientIP := net.ParseIP(realIP)
			if clientIP == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !trustedNet.Contains(clientIP) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
