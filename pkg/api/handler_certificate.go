package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

func (h *Handler) getCertificates(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

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

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.Error().Err(err).Msg("Unable to encode certificates")
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) getCertificate(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	certID := mux.Vars(request)["certificateID"]

	if h.tlsManager == nil {
		writeError(rw, fmt.Sprintf("certificate not found: %s", certID), http.StatusNotFound)
		return
	}

	certs := h.tlsManager.GetServerCertificates()
	x509Cert, ok := certs[certID]

	if !ok {
		writeError(rw, fmt.Sprintf("certificate not found: %s", certID), http.StatusNotFound)
		return
	}

	cert := buildCertificateRepresentation(x509Cert)

	if err := json.NewEncoder(rw).Encode(cert); err != nil {
		log.Error().Err(err).Str("id", certID).Msg("Unable to encode certificate")
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) extractCertificates() []certificateRepresentation {
	if h.tlsManager == nil {
		return []certificateRepresentation{}
	}

	x509Certs := h.tlsManager.GetServerCertificates()
	result := make([]certificateRepresentation, 0, len(x509Certs))

	for _, cert := range x509Certs {
		rep := buildCertificateRepresentation(cert)
		result = append(result, rep)
	}

	return result
}
