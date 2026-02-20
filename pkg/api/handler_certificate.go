package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func keepCertificate(cert certificateRepresentation, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	// Combine all searchable fields
	searchFields := make([]string, 0, 3+len(cert.SANs))
	searchFields = append(searchFields, cert.CommonName, cert.IssuerOrg, cert.IssuerCN)
	searchFields = append(searchFields, cert.SANs...)

	return criterion.withStatus(cert.Status) &&
		criterion.searchIn(searchFields...)
}

func (h Handler) getCertificates(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	if h.runtimeConfiguration == nil {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprint(rw, "[]")
		return
	}

	allCerts := h.extractCertificates()

	query := request.URL.Query()
	criterion := newSearchCriterion(query)

	results := make([]certificateRepresentation, 0, len(allCerts))
	for _, cert := range allCerts {
		if keepCertificate(cert, criterion) {
			results = append(results, cert)
		}
	}

	sortCertificates(query, results)

	err := json.NewEncoder(rw).Encode(results)
	if err != nil {
		log.Error().Err(err).Msg("Unable to encode certificates")
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getCertificate(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	certKey := vars["certKey"]

	// Decode certKey to get domains
	domains, err := getDomainsFromURLEncodedCertKey(certKey)
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Get certificate directly from store using domains
	store := h.tlsManager.GetStore("default")
	if store == nil {
		writeError(rw, "TLS store not found", http.StatusNotFound)
		return
	}

	certData := store.GetCertificate(domains)
	if certData == nil {
		writeError(rw, fmt.Sprintf("certificate not found: %s", certKey), http.StatusNotFound)
		return
	}

	// Get x509 certificate from Leaf
	x509Cert := certData.Certificate.Leaf

	cert := buildCertificateRepresentation(x509Cert)

	if err := json.NewEncoder(rw).Encode(cert); err != nil {
		log.Error().Err(err).Str("certKey", certKey).Msg("Unable to encode certificate")
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) extractCertificates() []certificateRepresentation {
	if h.tlsManager == nil {
		return []certificateRepresentation{}
	}

	x509Certs := h.tlsManager.GetServerCertificates()
	result := make([]certificateRepresentation, 0, len(x509Certs))

	for _, cert := range x509Certs {
		rep := buildCertificateRepresentation(cert)
		result = append(result, rep)
	}

	// Sort by commonName
	sort.Slice(result, func(i, j int) bool {
		return result[i].CommonName < result[j].CommonName
	})

	return result
}
