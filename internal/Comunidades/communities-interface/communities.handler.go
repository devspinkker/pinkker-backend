package communitiestinterfaces

import (
	communitiesapplication "PINKKER-BACKEND/internal/Comunidades/communities-application"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommunitiesHandler struct {
	Service *communitiesapplication.CommunitiesService
}

func NewCommunitiesHandler(service *communitiesapplication.CommunitiesService) *CommunitiesHandler {
	return &CommunitiesHandler{
		Service: service,
	}
}

func (h *CommunitiesHandler) CreateCommunity(c *fiber.Ctx) error {
	var req struct {
		CommunityName string   `json:"community_name"`
		Description   string   `json:"description"`
		IsPrivate     bool     `json:"is_private"`
		Categories    []string `json:"categories"`
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}

	community, err := h.Service.CreateCommunity(c.Context(), req.CommunityName, idValueToken, req.Description, req.IsPrivate, req.Categories)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error creating community",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Community created successfully",
		"community": community,
	})
}
func (h *CommunitiesHandler) AddMember(c *fiber.Ctx) error {
	var req struct {
		CommunityID primitive.ObjectID `json:"community_id"`
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}

	err := h.Service.AddMember(c.Context(), req.CommunityID, idValueToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error adding member",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Member added successfully",
	})
}
func (h *CommunitiesHandler) RemoveMember(c *fiber.Ctx) error {
	var req struct {
		CommunityID primitive.ObjectID `json:"community_id"`
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}

	err := h.Service.RemoveMember(c.Context(), req.CommunityID, idValueToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error adding member",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Member added successfully",
	})
}

func (h *CommunitiesHandler) BanMember(c *fiber.Ctx) error {
	var req struct {
		CommunityID primitive.ObjectID `json:"community_id"`
		UserID      primitive.ObjectID `json:"user_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}

	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	err := h.Service.BanMember(c.Context(), req.CommunityID, req.UserID, idValueToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error banning member",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Member banned successfully",
	})
}
func (h *CommunitiesHandler) GetCommunityPosts(c *fiber.Ctx) error {
	var req struct {
		CommunityIDs     primitive.ObjectID   `json:"community_ids"`
		ExcludeFilterIDs []primitive.ObjectID `json:"ExcludeFilterIDs"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	// Llamar al servicio para obtener los posts de las comunidades
	posts, err := h.Service.GetCommunityPosts(c.Context(), req.CommunityIDs, req.ExcludeFilterIDs, idValueToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching community posts",
			"error":   err.Error(),
		})
	}

	// Devolver los posts como respuesta JSON
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Posts fetched successfully",
		"posts":   posts,
	})
}
func (h *CommunitiesHandler) AddModerator(c *fiber.Ctx) error {
	var req struct {
		CommunityID string `json:"community_id"`
		NewModID    string `json:"new_mod_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}

	communityID, err := primitive.ObjectIDFromHex(req.CommunityID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid community ID",
		})
	}

	newModID, err := primitive.ObjectIDFromHex(req.NewModID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid new moderator ID",
		})
	}

	creatorID := c.Context().UserValue("_id").(string)
	creatorObjectID, err := primitive.ObjectIDFromHex(creatorID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	err = h.Service.AddModerator(c.Context(), communityID, newModID, creatorObjectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error adding moderator",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Moderator added successfully",
	})
}
func (h *CommunitiesHandler) DeletePost(c *fiber.Ctx) error {
	var req struct {
		CommunityID primitive.ObjectID `json:"CommunityID"`
		PostId      primitive.ObjectID `json:"PostId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}

	IdToken := c.Context().UserValue("_id").(string)
	IdTokenObjectID, err := primitive.ObjectIDFromHex(IdToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	err = h.Service.DeletePost(c.Context(), req.CommunityID, req.PostId, IdTokenObjectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Delete Post successfully",
	})
}
func (h *CommunitiesHandler) DeleteCommunity(c *fiber.Ctx) error {
	var req struct {
		CommunityID primitive.ObjectID `json:"CommunityID"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}

	IdToken := c.Context().UserValue("_id").(string)
	IdTokenObjectID, err := primitive.ObjectIDFromHex(IdToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	err = h.Service.DeleteCommunity(c.Context(), req.CommunityID, IdTokenObjectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Delete Community successfully",
	})
}
func (h *CommunitiesHandler) FindCommunityByName(c *fiber.Ctx) error {
	communityName := c.Query("community")

	if communityName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Community name is required",
		})
	}

	community, err := h.Service.FindCommunityByName(c.Context(), communityName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully",
		"data":    community,
	})
}

func (h *CommunitiesHandler) GetTop10CommunitiesByMembers(c *fiber.Ctx) error {
	community, err := h.Service.GetTop10CommunitiesByMembers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully",
		"data":    community,
	})
}

func (h *CommunitiesHandler) GetCommunity(c *fiber.Ctx) error {
	communityIDStr := c.Query("community")
	communityID, err := primitive.ObjectIDFromHex(communityIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid community ID",
		})
	}

	community, err := h.Service.GetCommunity(c.Context(), communityID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully",
		"data":    community,
	})
}
func (h *CommunitiesHandler) GetCommunityWithUserMembership(c *fiber.Ctx) error {
	communityIDStr := c.Query("community")
	communityID, err := primitive.ObjectIDFromHex(communityIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid community ID",
		})
	}

	IdToken := c.Context().UserValue("_id").(string)
	IdTokenObjectID, err := primitive.ObjectIDFromHex(IdToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	community, err := h.Service.GetCommunityWithUserMembership(c.Context(), communityID, IdTokenObjectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully",
		"data":    community,
	})
}