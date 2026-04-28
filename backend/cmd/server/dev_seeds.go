package main

import (
	"context"
	"fmt"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	userRepo "github.com/Tangyd893/WorkPal/backend/internal/user/repo"
	"golang.org/x/crypto/bcrypt"
)

type developmentUserSeed struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
}

var developmentUserSeeds = []developmentUserSeed{
	{
		Username: "admin",
		Password: "admin123",
		Nickname: "Administrator",
		Email:    "admin@workpal.local",
		Phone:    "13800000000",
	},
	{
		Username: "emma.chen",
		Password: "workpal123",
		Nickname: "Emma Chen",
		Email:    "emma.chen@workpal.local",
		Phone:    "13800000001",
	},
	{
		Username: "liam.wang",
		Password: "workpal123",
		Nickname: "Liam Wang",
		Email:    "liam.wang@workpal.local",
		Phone:    "13800000002",
	},
	{
		Username: "sofia.zhao",
		Password: "workpal123",
		Nickname: "Sofia Zhao",
		Email:    "sofia.zhao@workpal.local",
		Phone:    "13800000003",
	},
}

func ensureDevelopmentUsers(ctx context.Context, userRepoInst *userRepo.UserRepo) error {
	for _, seed := range developmentUserSeeds {
		if err := ensureDevelopmentUser(ctx, userRepoInst, seed); err != nil {
			return err
		}
	}

	return nil
}

func ensureDevelopmentUser(ctx context.Context, userRepoInst *userRepo.UserRepo, seed developmentUserSeed) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash development password for %s: %w", seed.Username, err)
	}

	existingUser, err := userRepoInst.GetByUsername(ctx, seed.Username)
	if err != nil {
		if !apperrors.Is(err, apperrors.ErrUserNotFound) {
			return err
		}

		return userRepoInst.Create(ctx, &model.User{
			Username:     seed.Username,
			PasswordHash: string(passwordHash),
			Nickname:     seed.Nickname,
			Email:        seed.Email,
			Phone:        seed.Phone,
			Status:       1,
		})
	}

	existingUser.PasswordHash = string(passwordHash)
	existingUser.Nickname = seed.Nickname
	existingUser.Email = seed.Email
	existingUser.Phone = seed.Phone
	existingUser.Status = 1

	return userRepoInst.Update(ctx, existingUser)
}
