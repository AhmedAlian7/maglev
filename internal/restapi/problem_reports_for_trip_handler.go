package restapi

import (
	"log/slog"
	"net/http"

	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

func (api *RestAPI) problemReportsForTripHandler(w http.ResponseWriter, r *http.Request) {
	logger := api.Logger
	if logger == nil {
		logger = slog.Default()
	}

	compositeID := utils.ExtractIDFromParams(r)

	if err := utils.ValidateID(compositeID); err != nil {
		fieldErrors := map[string][]string{
			"id": {err.Error()},
		}
		api.validationErrorResponse(w, r, fieldErrors)
		return
	}

	_, tripID, err := utils.ExtractAgencyIDAndCodeID(compositeID)
	if err != nil {
		logger.Warn("problem reports for trip failed: invalid tripID format",
			slog.String("tripID", compositeID),
			slog.Any("error", err))
		api.sendError(w, r, http.StatusBadRequest, "invalid tripID format")
		return
	}

	// Safety check: Ensure DB is initialized
	if api.GtfsManager == nil || api.GtfsManager.GtfsDB == nil || api.GtfsManager.GtfsDB.Queries == nil {
		logger.Error("problem reports for trip failed: GTFS DB not initialized")
		api.sendError(w, r, http.StatusInternalServerError, "internal server error")
		return
	}

	ctx := r.Context()
	reports, err := api.GtfsManager.GtfsDB.Queries.GetProblemReportsByTrip(ctx, tripID)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	reportList := make([]models.ProblemReportTrip, 0, len(reports))
	for _, report := range reports {
		reportList = append(reportList, models.NewProblemReportTrip(report))
	}

	references := models.NewEmptyReferences()
	response := models.NewListResponse(reportList, references, false, api.Clock)
	api.sendResponse(w, r, response)
}
