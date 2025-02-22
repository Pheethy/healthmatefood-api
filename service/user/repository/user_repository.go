package repository

import (
	"context"
	"encoding/json"
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

func (u *userRepository) FetchOneUserByEmail(ctx context.Context, email string) (*models.UserSign, error) {
	sql := `
    SELECT
      to_jsonb("json_data")
    FROM (
      SELECT
        "users"."id",
        "users"."username",
        "users"."password",
        "users"."email",
        "roles"."name" "role",
        to_char("users"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
        to_char("users"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at",
        (
          SELECT
            COALESCE(array_to_json(array_agg("IM")), '[]'::json)
          FROM (
            SELECT
              "images"."id",
              "images"."filename",
              "images"."url",
              "images"."ref_id",
              "images"."ref_type",
              to_char("users"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
              to_char("users"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at"
            FROM
              "images"
            WHERE
              "images"."ref_id" = "users"."id"
            AND
              "images"."ref_type" = 'USER'
          ) AS "IM"
        ) AS "images",
        (
          SELECT
            to_jsonb("INFO")
          FROM (
            SELECT
              "user_info"."id",
              "user_info"."user_id",
              "user_info"."firstname",
              "user_info"."lastname",
              "user_info"."gender",
              "user_info"."height",
              "user_info"."weight",
              "user_info"."target",
              "user_info"."target_weight",
              "user_info"."active_level",
              to_char("user_info"."dob", 'yyyy-MM-dd HH:mm:ss') "dob",
              to_char("user_info"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
              to_char("user_info"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at"
            FROM
              "user_info"
            INNER JOIN
              "users"
            ON
              "users"."id" = "user_info"."user_id"
          ) AS "INFO"
        ) AS "user_info"
      FROM
        "users"
      INNER JOIN
        "roles"
      ON
        "users"."role_id" = "roles"."id"
      WHERE
        "users"."email" = $1::text
    ) AS "json_data"
  `
	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var jsonData []byte
	err = stmt.QueryRowxContext(ctx, email).Scan(&jsonData)
	if err != nil {
		return nil, err
	}

	user := new(models.UserSign)
	if err := json.Unmarshal(jsonData, &user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userRepository) FetchOneUserById(ctx context.Context, id *uuid.UUID) (*models.UserSign, error) {
	sql := `
    SELECT
      to_jsonb("json_data")
    FROM (
      SELECT
        "users"."id",
        "users"."username",
        "users"."email",
        "roles"."name" "role",
        to_char("users"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
        to_char("users"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at",
        (
          SELECT
            COALESCE(array_to_json(array_agg("IM")), '[]'::json)
          FROM (
            SELECT
              "images"."id",
              "images"."filename",
              "images"."url",
              "images"."ref_id",
              "images"."ref_type",
              to_char("users"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
              to_char("users"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at"
            FROM
              "images"
            WHERE
              "images"."ref_id" = "users"."id"
            AND
              "images"."ref_type" = 'USER'
          ) AS "IM"
        ) AS "images",
        (
          SELECT
            to_jsonb("INFO")
          FROM (
            SELECT
              "user_info"."id",
              "user_info"."user_id",
              "user_info"."firstname",
              "user_info"."lastname",
              "user_info"."gender",
              "user_info"."height",
              "user_info"."weight",
              "user_info"."target",
              "user_info"."target_weight",
              "user_info"."active_level",
              to_char("user_info"."dob", 'yyyy-MM-dd HH:mm:ss') "dob",
              to_char("user_info"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
              to_char("user_info"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at"
            FROM
              "user_info"
            INNER JOIN
              "users"
            ON
              "users"."id" = "user_info"."user_id"
          ) AS "INFO"
        ) AS "user_info"
      FROM
        "users"
      INNER JOIN
        "roles"
      ON
        "users"."role_id" = "roles"."id"
      WHERE
        "users"."id" = $1::uuid
    ) AS "json_data"
  `
	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var jsonData []byte
	err = stmt.QueryRowxContext(ctx, id).Scan(&jsonData)
	if err != nil {
		return nil, err
	}

	user := new(models.UserSign)
	if err := json.Unmarshal(jsonData, &user); err != nil {
		return nil, err
	}
	user.UserInfo.GetBMR()

	return user, nil
}

func (u *userRepository) FetchAllUsers(ctx context.Context, args *sync.Map) ([]*models.User, error) {
	sql := `
    SELECT
      COALESCE(array_to_json(array_agg("json_data")), '[]'::json)
    FROM (
      SELECT
        "users"."id",
        "users"."username",
        "users"."email",
        "roles"."name" "role",
        to_char("users"."created_at", 'yyyy-MM-dd HH:mm:ss') AS "created_at",
        to_char("users"."updated_at", 'yyyy-MM-dd HH:mm:ss') AS "updated_at",
        (
          SELECT
            COALESCE(array_to_json(array_agg("IM")), '[]'::json)
          FROM (
            SELECT
              "images"."id",
              "images"."filename",
              "images"."url",
              "images"."ref_id",
              "images"."ref_type",
              to_char("images"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
              to_char("images"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at"
            FROM
              "images"
            WHERE
              "images"."ref_id" = "users"."id"
            AND
              "images"."ref_type" = 'USER'
          ) AS "IM"
        ) AS "images",
        (
          SELECT
            to_jsonb("INFO")
          FROM (
            SELECT
              "user_info"."id",
              "user_info"."user_id",
              "user_info"."firstname",
              "user_info"."lastname",
              "user_info"."gender",
              "user_info"."height",
              "user_info"."weight",
              "user_info"."target",
              "user_info"."target_weight",
              "user_info"."active_level",
              to_char("user_info"."dob", 'yyyy-MM-dd HH:mm:ss') "dob",
              to_char("user_info"."created_at", 'yyyy-MM-dd HH:mm:ss') "created_at",
              to_char("user_info"."updated_at", 'yyyy-MM-dd HH:mm:ss') "updated_at"
            FROM
              "user_info"
            WHERE
              "users"."id" = "user_info"."user_id"
          ) AS "INFO"
        ) AS "user_info"
      FROM
        "users"
      LEFT JOIN
        "roles"
      ON
        "users"."role_id" = "roles"."id"
    ) AS "json_data"
  `

	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var jsonData []byte
	err = stmt.QueryRowxContext(ctx).Scan(&jsonData)
	if err != nil {
		return nil, err
	}

	users := make([]*models.User, 0)
	if err := json.Unmarshal(jsonData, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *userRepository) FetchOneOAuthByRefreshToken(ctx context.Context, refreshToken string) (*models.OAuth, error) {
	sql := `
    SELECT
      to_josnb("json_data")
    FROM (
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
    ) AS "json_data"
	`

	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	jsonData := make([]byte, 0)
	if err = stmt.QueryRowxContext(ctx, refreshToken).Scan(&jsonData); err != nil {
		return nil, err
	}

	oauth := new(models.OAuth)
	if err := json.Unmarshal(jsonData, &oauth); err != nil {
		return nil, err
	}

	return oauth, nil
}

func (u *userRepository) FetchOneUserInfoByUserId(ctx context.Context, userId *uuid.UUID) (*models.UserInfo, error) {
	sql := `
    SELECT
      to_jsonb("json_data")
    FROM (
      SELECT
        "user_info"."id",
        "user_info"."user_id",
        "user_info"."age",
        "user_info"."gender",
        "user_info"."height",
        "user_info"."weight",
        "user_info"."target_weight",
        "user_info"."active_level",
        "user_info"."created_at",
        "user_info"."updated_at"
      FROM
        "user_info"
      WHERE
        "user_info"."user_id" = $1::uuid
    ) AS "json_data"
  `

	stmt, err := u.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var jsonData []byte
	err = stmt.QueryRowxContext(ctx, userId).Scan(&jsonData)
	if err != nil {
		return nil, err
	}

	userInfo := new(models.UserInfo)
	if err := json.Unmarshal(jsonData, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
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
      "email",
      "role_id",
      "created_at",
      "updated_at"
    ) VALUES (
      $1::uuid,
      $2::text,
      $3::text,
      $4::text,
      $5::int,
      $6::timestamp,
      $7::timestamp
    )
		ON CONFLICT (id)
		DO UPDATE SET
      email=$8::text,
      password=$9::text,
      updated_at=$10::timestamp
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
		user.Email,
		user.RoleId,
		user.CreatedAt,
		user.UpdatedAt,
		/* Update */
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

func (u *userRepository) UpsertUserInfo(ctx context.Context, userInfo *models.UserInfo) error {
	tx, err := u.psqlDB.Beginx()
	if err != nil {
		return err
	}
	sql := `
    INSERT INTO "user_info" (
      "id",
      "user_id",
      "firstname",
      "lastname",
      "gender",
      "height",
      "weight",
      "target",
      "target_weight",
      "active_level",
      "dob",
      "created_at",
      "updated_at"
    ) VALUES (
      $1::uuid,
      $2::uuid,
      $3::text,
      $4::text,
      $5::gender_type,
      $6::float,
      $7::float,
      $8::target_type,
      $9::float,
      $10::active_level_type,
      $11::timestamp,
      $12::timestamp,
      $13::timestamp
    )
	ON CONFLICT (id)
	DO UPDATE SET
      firstname=$14::text,
      lastname=$15::text,
      gender=$16::gender_type,
      height=$17::float,
      weight=$18::float,
      target=$19::target_type,
      target_weight=$20::float,
      dob=$21::timestamp,
      updated_at=$22::timestamp
  `
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		/* Create */
		userInfo.Id,
		userInfo.UserId,
		userInfo.Firstname,
		userInfo.Lastname,
		userInfo.Gender,
		userInfo.Height,
		userInfo.Weight,
		userInfo.Weight,
		userInfo.Target,
		userInfo.TargetWeight,
		userInfo.ActiveLevel,
		userInfo.DOB,
		userInfo.CreatedAt,
		userInfo.UpdatedAt,
		/* Update */
		userInfo.Firstname,
		userInfo.Lastname,
		userInfo.Gender,
		userInfo.Height,
		userInfo.Weight,
		userInfo.Weight,
		userInfo.Target,
		userInfo.TargetWeight,
		userInfo.ActiveLevel,
		userInfo.DOB,
		userInfo.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
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

func (u *userRepository) ormOneUserInfo(ctx context.Context, rows *sqlx.Rows) (*models.UserInfo, error) {
	mapping, err := orm.OrmContext(ctx, new(models.UserInfo), rows, orm.NewMapperOption())
	if err != nil {
		return nil, err
	}
	userInfo := mapping.GetData().([]*models.UserInfo)
	if len(userInfo) == 0 {
		return nil, errors.New(constants.ERROR_OAUTH_NOT_FOUND)
	}
	return userInfo[0], nil
}
