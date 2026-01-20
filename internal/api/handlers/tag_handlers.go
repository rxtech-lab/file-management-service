package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// ListTags implements generated.StrictServerInterface
func (h *StrictHandlers) ListTags(
	ctx context.Context,
	request generated.ListTagsRequestObject,
) (generated.ListTagsResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListTags401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	keyword := deref(request.Params.Keyword)
	limit := derefInt(request.Params.Limit, 100)
	offset := derefInt(request.Params.Offset, 0)

	tags, total, err := h.tagService.ListTags(userID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	return generated.ListTags200JSONResponse{
		Data:   tagListToGenerated(tags),
		Total:  int(total),
		Limit:  limit,
		Offset: offset,
	}, nil
}

// CreateTag implements generated.StrictServerInterface
func (h *StrictHandlers) CreateTag(
	ctx context.Context,
	request generated.CreateTagRequestObject,
) (generated.CreateTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.CreateTag400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	tag := &models.Tag{
		Name:        request.Body.Name,
		Color:       deref(request.Body.Color),
		Description: deref(request.Body.Description),
	}

	if err := h.tagService.CreateTag(userID, tag); err != nil {
		return generated.CreateTag400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.CreateTag201JSONResponse(tagModelToGenerated(tag)), nil
}

// GetTag implements generated.StrictServerInterface
func (h *StrictHandlers) GetTag(
	ctx context.Context,
	request generated.GetTagRequestObject,
) (generated.GetTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	tag, err := h.tagService.GetTagByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return generated.GetTag404JSONResponse{NotFoundJSONResponse: notFound("Tag not found")}, nil
	}

	return generated.GetTag200JSONResponse(tagModelToGenerated(tag)), nil
}

// UpdateTag implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateTag(
	ctx context.Context,
	request generated.UpdateTagRequestObject,
) (generated.UpdateTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.UpdateTag400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Get existing tag
	existing, err := h.tagService.GetTagByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return generated.UpdateTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Update fields
	if request.Body.Name != nil {
		existing.Name = *request.Body.Name
	}
	if request.Body.Color != nil {
		existing.Color = *request.Body.Color
	}
	if request.Body.Description != nil {
		existing.Description = *request.Body.Description
	}

	if err := h.tagService.UpdateTag(userID, existing); err != nil {
		return generated.UpdateTag400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated tag
	updated, err := h.tagService.GetTagByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.UpdateTag200JSONResponse(tagModelToGenerated(updated)), nil
}

// DeleteTag implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteTag(
	ctx context.Context,
	request generated.DeleteTagRequestObject,
) (generated.DeleteTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.tagService.DeleteTag(userID, uint(request.Id)); err != nil {
		return generated.DeleteTag404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.DeleteTag204Response{}, nil
}
