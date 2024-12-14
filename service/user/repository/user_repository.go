package repository

import (
	"context"
	"errors"
	"fmt"
	"healthmatefood-api/constants"
	"healthmatefood-api/models"
	"healthmatefood-api/service/user"
	"strings"
	"sync"

	"github.com/Pheethy/psql/orm"
	"github.com/Pheethy/sqlx"
	"github.com/gofrs/uuid"
)

type userRepository struct {
	psqlDB *sqlx.DB
}

func NewUserRepository(psqlDB *sqlx.DB) user.IUserRepository {
	return &userRepository{
		psqlDB: psqlDB,
	}
}

func (u *userRepository) FetchOneUserByEmail(ctx context.Context, email string) (*models.User, error) {
	sql := fmt.Sprintf(`
    SELECT
      "users"."id",
      "users"."username",
      "users"."password",
      "users"."firstname",
      "users"."lastname",
      "users"."email",
      "roles"."name" "role",
      "users"."created_at",
      "users"."updated_at",
      %s
    FROM
      "users"
    INNER JOIN
      "roles"
    ON
      "users"."role_id" = "roles"."id"
    LEFT JOIN
      "images"
    ON
      "images"."ref_id" = "users"."id"
    AND
      "images"."ref_type" = 'USER'
    WHERE
      "users"."email" = $1::text
  `, orm.GetSelector(models.Image{}))
	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := u.ormOneUser(ctx, rows)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userRepository) FetchOneUserById(ctx context.Context, id *uuid.UUID) (*models.User, error) {
	sql := fmt.Sprintf(`
    SELECT
      "users"."id",
      "users"."username",
      "users"."password",
      "users"."firstname",
      "users"."lastname",
      "users"."email",
      "roles"."name" "role",
      "users"."created_at",
      "users"."updated_at",
      %s
    FROM
      "users"
    INNER JOIN
      "roles"
    ON
      "users"."role_id" = "roles"."id"
    LEFT JOIN
      "images"
    ON
      "images"."ref_id" = "users"."id"
    AND
      "images"."ref_type" = 'USER'
    WHERE
      "users"."id" = $1::uuid
  `, orm.GetSelector(models.Image{}))
	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := u.ormOneUser(ctx, rows)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userRepository) FetchAllUsers(ctx context.Context, args *sync.Map) ([]*models.User, error) {
	sql := fmt.Sprintf(`
    SELECT
      "users"."id",
      "users"."username",
      "users"."password",
      "users"."email",
      "roles"."name" "role",
      "users"."created_at",
      "users"."updated_at",
      %s
    FROM
      "users"
    INNER JOIN
      "roles"
    ON
      "users"."role_id" = "roles"."id"
    INNER JOIN
      "images"
    ON
      "images"."ref_id" = "users"."id"
    WHERE
      "images"."ref_type" = 'USER'
  `, orm.GetSelector(models.Image{}))

	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := u.ormAllUser(ctx, rows)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u *userRepository) FetchOneOAuthByRefreshToken(ctx context.Context, refreshToken string) (*models.OAuth, error) {
	sql := `
	SELECT
		"oauth"."id",
		"oauth"."user_id",
		"oauth"."access_token",
		"oauth"."refresh_token",
		"oauth"."created_at",
		"oauth"."updated_at"
	FROM
		"oauth"
	WHERE
		"oauth"."refresh_token" = $1::text
	`
	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	oauth, err := u.ormOneOAuth(ctx, rows)
	if err != nil {
		return nil, err
	}

	return oauth, nil
}

func (u *userRepository) UpsertImages(ctx context.Context, user *models.User) error {
	tx, err := u.psqlDB.Beginx()
	if err != nil {
		return err
	}
	sql := `
		INSERT INTO images (
	        id,
	        filename,
	        url,
	        ref_id,
	        ref_type,
	        created_at,
	        updated_at
		) VALUES (
	        $1::uuid,
	        $2::text,
	        $3::text,
	        $4::uuid,
	        $5::image_ref_type,
	        $6::timestamp,
	        $7::timestamp
		)
		ON CONFLICT (id)
		DO UPDATE SET
	        filename=$8::text,
	        url=$9::text,
	        ref_id=$10::uuid,
	        ref_type=$11::image_ref_type,
	        updated_at=$12::timestamp
	`
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("prepare failed: %v", err.Error())
	}

	if len(user.Images) > 0 {
		for index := range user.Images {
			if _, err := stmt.ExecContext(ctx,
				// create
				user.Images[index].Id,
				user.Images[index].FileName,
				user.Images[index].URL,
				user.Id,
				user.Images[index].RefType,
				user.Images[index].CreatedAt,
				user.Images[index].UpdatedAt,
				// update
				user.Images[index].FileName,
				user.Images[index].URL,
				user.Id,
				user.Images[index].RefType,
				user.Images[index].UpdatedAt,
			); err != nil {
				tx.Rollback()
				return fmt.Errorf("exec failed: %v", err)
			}
		}
	}

	return tx.Commit()
}

func (u *userRepository) UpsertUser(ctx context.Context, user *models.User) error {
	tx, err := u.psqlDB.Beginx()
	if err != nil {
		return err
	}
	sql := `
    INSERT INTO "users" (
      "id",
      "username",
      "password",
      "firstname",
      "lastname",
      "email",
      "role_id",
      "created_at",
      "updated_at"
    ) VALUES (
      $1::uuid,
      $2::text,
      $3::text,
      $4::text,
      $5::text,
      $6::text,
      $7::int,
      $8::timestamp,
      $9::timestamp
    )
		ON CONFLICT (id)
		DO UPDATE SET
      firstname=$10::text,
      lastname=$11::text,
      email=$12::text,
      password=$13::text,
      updated_at=$14::timestamp
  `
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		/* Create */
		user.Id,
		user.Username,
		user.Password,
		user.Firstname,
		user.Lastname,
		user.Email,
		user.RoleId,
		user.CreatedAt,
		user.UpdatedAt,
		/* Update */
		user.Firstname,
		user.Lastname,
		user.Email,
		user.Password,
		user.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		if ok := strings.Contains(err.Error(), constants.POSTGRES_ERROR_USERNAME_WAS_DUPLICATED); ok {
			return errors.New(constants.ERROR_USERNAME_WAS_DUPLICATED)
		}
		if ok := strings.Contains(err.Error(), constants.POSTGRES_ERROR_EMAIL_WAS_DUPLICATED); ok {
			return errors.New(constants.ERROR_EMAIL_WAS_DUPLICATED)
		}
		return err
	}
	return tx.Commit()
}

func (u *userRepository) UpsertOAuth(ctx context.Context, oauth *models.OAuth) error {
	tx, err := u.psqlDB.Beginx()
	if err != nil {
		return err
	}
	sql := `
	    INSERT INTO "oauth" (
	      "id",
	      "user_id",
	      "access_token",
	      "refresh_token",
	      "created_at",
	      "updated_at"
	    ) VALUES (
	      $1::uuid,
	      $2::uuid,
	      $3::text,
	      $4::text,
	      $5::timestamp,
	      $6::timestamp
	    )
		ON CONFLICT (id)
		DO UPDATE SET
	      access_token=$7::text,
	      refresh_token=$8::text,
	      updated_at=$9::timestamp
  `
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		/* Create */
		oauth.Id,
		oauth.UserId,
		oauth.AccessToken,
		oauth.RefreshToken,
		oauth.CreatedAt,
		oauth.UpdatedAt,
		/* Update */
		oauth.AccessToken,
		oauth.RefreshToken,
		oauth.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (u *userRepository) ormOneUser(ctx context.Context, rows *sqlx.Rows) (*models.User, error) {
	mapping, err := orm.OrmContext(ctx, new(models.User), rows, orm.NewMapperOption())
	if err != nil {
		return nil, err
	}
	users := mapping.GetData().([]*models.User)
	if len(users) == 0 {
		return nil, errors.New(constants.ERROR_USER_NOT_FOUND)
	}
	return users[0], nil
}

func (u *userRepository) ormAllUser(ctx context.Context, rows *sqlx.Rows) ([]*models.User, error) {
	mapping, err := orm.OrmContext(ctx, new(models.User), rows, orm.NewMapperOption())
	if err != nil {
		return nil, err
	}
	users := mapping.GetData().([]*models.User)
	if len(users) == 0 {
		return nil, errors.New(constants.ERROR_USER_NOT_FOUND)
	}
	return users, nil
}

func (u *userRepository) ormOneOAuth(ctx context.Context, rows *sqlx.Rows) (*models.OAuth, error) {
	mapping, err := orm.OrmContext(ctx, new(models.OAuth), rows, orm.NewMapperOption())
	if err != nil {
		return nil, err
	}
	oauths := mapping.GetData().([]*models.OAuth)
	if len(oauths) == 0 {
		return nil, errors.New(constants.ERROR_OAUTH_NOT_FOUND)
	}
	return oauths[0], nil
}
