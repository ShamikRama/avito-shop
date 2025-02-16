//go:build integration
// +build integration

package integrations

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"avito-shop/internal/repository"
	"avito-shop/internal/service"
	"context"
	"database/sql"
	"fmt"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

type IntegrationTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo service.RepoUserInterface
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5433/shopTest?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to reach database: %v", err)
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		fmt.Println("goose dialect is not set")
	}

	err = goose.Up(db, "../../migrations")
	if err != nil {
		fmt.Println("migrations for test db is not up")
	}

	s.db = db
}

func (s *IntegrationTestSuite) TearDownSuite() {
	err := goose.Reset(s.db, "../../migrations")
	if err != nil {
		fmt.Println("migrations test db did not reset")
	}

	err = s.db.Close()
	if err != nil {
		fmt.Println("test db is not closed")
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	_, err := s.db.Exec(`TRUNCATE TABLE 
            users, 
            items, 
            purchases,
    		transfers
        RESTART IDENTITY CASCADE`)
	if err != nil {
		s.FailNow("Failed to clean tables", err.Error())
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	log := logger.NewLogger()
	s.repo = repository.NewUserRepo(s.db, log)
}

func (s *IntegrationTestSuite) saveTestUser(user domain.User) int {
	query := `INSERT INTO users (username, password_hash, balance) 
              VALUES ($1, $2, $3) 
              RETURNING id`

	var id int
	err := s.db.QueryRow(query,
		user.Username,
		user.PasswordHash,
		1000,
	).Scan(&id)

	if err != nil {
		s.Fail(err.Error())
	}

	return id
}

func (s *IntegrationTestSuite) saveTestItem(itemName string) int {
	query := `INSERT INTO items (name, price) 
              VALUES ($1, $2) 
              RETURNING id`

	var id int
	err := s.db.QueryRow(query,
		itemName,
		20,
	).Scan(&id)

	if err != nil {
		s.Fail(err.Error())
	}

	return id
}

func (s *IntegrationTestSuite) TestGetUser() {
	existUser := domain.User{
		Username:     "testUser1",
		PasswordHash: "testUser1password",
		Coins:        1000,
	}

	id := s.saveTestUser(existUser)

	user, err := s.repo.GetUser(context.Background(), id)

	s.Require().NoError(err)
	s.Require().NotEqual(0, user.ID)

	s.Require().Equal(id, user.ID)
	s.Require().Equal(user.Username, existUser.Username)
}

func (s *IntegrationTestSuite) TestGetNotFoundUser() {
	_, err := s.repo.GetUser(context.Background(), 2)

	s.Require().Error(err)
	s.Require().ErrorIs(err, erorrs.ErrNotFound)
}

func (s *IntegrationTestSuite) TestGetItem() {
	existItem := domain.Item{
		Name:  "cup",
		Price: 20,
	}

	id := s.saveTestItem(existItem.Name)

	item, err := s.repo.GetItem(context.Background(), "cup")

	s.Require().NoError(err)
	s.Require().NotEqual(0, item.ID)

	s.Require().Equal(id, item.ID)
	s.Require().Equal(item.Name, existItem.Name)
}

func (s *IntegrationTestSuite) TestGetNotFoundItem() {
	_, err := s.repo.GetItem(context.Background(), "car")

	s.Require().Error(err)
	s.Require().ErrorIs(err, erorrs.ErrItemNotFound)
}

func (s *IntegrationTestSuite) TestBuyItem_Success() {
	userID := s.saveTestUser(domain.User{
		Username:     "testUser2",
		PasswordHash: "testUser2password",
	})

	_ = s.saveTestItem("cup")

	userFromDB, err := s.repo.GetUser(context.Background(), userID)
	s.Require().NoError(err)

	itemFromDB, err := s.repo.GetItem(context.Background(), "cup")
	s.Require().NoError(err)

	err = s.repo.BuyItem(context.Background(), userFromDB, itemFromDB)
	s.Require().NoError(err)

	var balance int
	err = s.db.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	s.Require().NoError(err)
	s.Require().Equal(1000-itemFromDB.Price, balance)

	items, err := s.repo.GetPurchasedItems(context.Background(), userID)
	s.Require().NoError(err)
	s.Require().Len(items, 1)
	s.Require().Equal(1, items[0].Quantity)
}

func (s *IntegrationTestSuite) TestBuyItem_UserNotFound() {
	itemID := s.saveTestItem("item")
	item, err := s.repo.GetItem(context.Background(), "item")
	s.Require().NoError(err)

	ghostUser := domain.User{
		ID:       999,
		Username: "ghost",
	}

	err = s.repo.BuyItem(context.Background(), ghostUser, item)

	s.Require().Error(err)
	s.Require().ErrorIs(err, erorrs.ErrNotFound)

	var purchaseCount int
	err = s.db.QueryRow(
		"SELECT COUNT(*) FROM purchases WHERE item_id = $1",
		itemID,
	).Scan(&purchaseCount)
	s.Require().NoError(err)
	s.Require().Equal(0, purchaseCount)
}

func (s *IntegrationTestSuite) TestSendCoinToUser_InsufficientFunds() {
	fromUserID := s.saveTestUser(domain.User{
		Username:     "poor_sender",
		PasswordHash: "poor_pass",
	})

	toUserID := s.saveTestUser(domain.User{
		Username:     "rich_receiver",
		PasswordHash: "rich_pass",
	})

	err := s.repo.SendCoins(context.Background(), fromUserID, toUserID, 1500)
	s.Require().ErrorIs(err, erorrs.ErrInsufficientFunds)
}
