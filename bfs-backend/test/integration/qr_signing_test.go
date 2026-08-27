package integration

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/product"
	"backend/internal/qrsign"
	"backend/internal/service"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func decodePub(s string) (ed25519.PublicKey, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(b), nil
}

func TestQRKeyService_DerivesStableKeypair(t *testing.T) {
	svc := NewQRKeySvc()

	priv1, pub1 := svc.SigningKey()
	require.Len(t, pub1, ed25519.PublicKeySize)

	priv2, pub2 := svc.SigningKey()
	require.Equal(t, []byte(pub1), []byte(pub2))
	require.Equal(t, []byte(priv1), []byte(priv2))

	servedPub := svc.PublicKey()
	require.Equal(t, []byte(pub1), []byte(servedPub))

	tok, err := qrsign.Sign(priv1, qrsign.Payload{
		Version:   qrsign.Version,
		OrderID:   "o",
		IssuedAt:  1,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Lines:     []qrsign.Line{{ProductID: "p", Quantity: 1}},
	})
	require.NoError(t, err)
	_, err = qrsign.Verify(servedPub, tok, time.Now())
	require.NoError(t, err, "served public key must verify a token from the signing key")
}

func TestQRKeyService_RequiresSeed(t *testing.T) {
	// Empty seed must fail startup, not create unredeemable orders.
	_, err := service.NewQRKeyService(config.Config{}, zap.NewNop())
	require.Error(t, err)

	// Malformed (wrong-length) seed must fail too.
	_, err = service.NewQRKeyService(config.Config{
		QRSigning: config.QRSigningConfig{Ed25519PrivateSeed: "dG9vc2hvcnQ="},
	}, zap.NewNop())
	require.Error(t, err)
}

func TestPaymentService_SignsQRPayloadWhenPaid(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()
	svc := service.NewPaymentService(
		cfg,
		tdb.Client,
		repos.Order,
		repos.OrderLine,
		repos.OrderPayment,
		NewProductSvc(repos),
		repos.MenuSlot,
		repos.Inventory,
		nil,
		nil,
		NewQRKeySvc(),
		zap.NewNop(),
	)
	keys := NewQRKeySvc()
	ctx := context.Background()

	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)
	sprite := fixtures.CreateProduct("Sprite", category.ID, 350, product.TypeSimple, nil)
	fixtures.AddInventory(cola.ID, 100, inventoryledger.ReasonOpeningBalance)
	fixtures.AddInventory(sprite.ID, 100, inventoryledger.ReasonOpeningBalance)

	input := service.CreateCheckoutInput{
		Items: []service.CheckoutItemInput{
			{ProductID: cola.ID, Quantity: 2},
			{ProductID: sprite.ID, Quantity: 1},
		},
	}

	prep, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
	require.NoError(t, err)

	ord, err := repos.Order.GetByID(ctx, prep.OrderID)
	require.NoError(t, err)
	require.Nil(t, ord.QrPayload, "pending order must not carry a pickup token")

	empty, err := svc.EnsureOrderQRToken(ctx, prep.OrderID)
	require.NoError(t, err)
	require.Empty(t, empty, "no token is minted while the order is unpaid")

	require.NoError(t, svc.MarkOrderPaidDev(ctx, prep.OrderID))

	token, err := svc.EnsureOrderQRToken(ctx, prep.OrderID)
	require.NoError(t, err)
	require.NotEmpty(t, token, "a paid order must carry a signed qr payload")

	again, err := svc.EnsureOrderQRToken(ctx, prep.OrderID)
	require.NoError(t, err)
	require.Equal(t, token, again, "re-signing a paid order returns the stored token")

	ord, err = repos.Order.GetByID(ctx, prep.OrderID)
	require.NoError(t, err)
	require.NotNil(t, ord.QrPayload)
	require.Equal(t, token, *ord.QrPayload)

	pub := keys.PublicKey()

	payload, err := qrsign.Verify(pub, token, time.Now())
	require.NoError(t, err)
	require.Equal(t, qrsign.Version, payload.Version)
	require.Equal(t, prep.OrderID, payload.OrderID)
	require.NotZero(t, payload.IssuedAt)
	require.NotZero(t, payload.ExpiresAt)
	require.GreaterOrEqual(t, payload.ExpiresAt, payload.IssuedAt)

	_, err = qrsign.Verify(pub, token, time.Unix(payload.ExpiresAt, 0).Add(time.Second))
	require.ErrorIs(t, err, qrsign.ErrExpired)

	require.Len(t, payload.Lines, 2)
	byProduct := map[string]int{}
	for _, l := range payload.Lines {
		byProduct[l.ProductID] = l.Quantity
	}
	require.Equal(t, 2, byProduct[cola.ID])
	require.Equal(t, 1, byProduct[sprite.ID])
}

func TestPaymentService_SignsBundleComponentsWhenPaid(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()
	svc := service.NewPaymentService(
		cfg,
		tdb.Client,
		repos.Order,
		repos.OrderLine,
		repos.OrderPayment,
		NewProductSvc(repos),
		repos.MenuSlot,
		repos.Inventory,
		nil,
		nil,
		NewQRKeySvc(),
		zap.NewNop(),
	)
	keys := NewQRKeySvc()
	ctx := context.Background()

	cat := fixtures.CreateCategory("Food", 1, true)
	combo := fixtures.CreateProduct("Combo", cat.ID, 900, product.TypeMenu, nil)
	burger := fixtures.CreateProduct("Burger", cat.ID, 0, product.TypeSimple, nil)
	drink := fixtures.CreateProduct("Drink", cat.ID, 0, product.TypeSimple, nil)
	fixtures.AddInventory(burger.ID, 50, inventoryledger.ReasonOpeningBalance)
	fixtures.AddInventory(drink.ID, 50, inventoryledger.ReasonOpeningBalance)

	mainSlot := fixtures.CreateMenuSlot(combo.ID, "Main", 1)
	drinkSlot := fixtures.CreateMenuSlot(combo.ID, "Drink", 2)
	fixtures.CreateMenuSlotOption(mainSlot.ID, burger.ID)
	fixtures.CreateMenuSlotOption(drinkSlot.ID, drink.ID)

	input := service.CreateCheckoutInput{
		Items: []service.CheckoutItemInput{
			{
				ProductID: combo.ID,
				Quantity:  2,
				Configuration: map[string]string{
					mainSlot.ID:  burger.ID,
					drinkSlot.ID: drink.ID,
				},
			},
		},
	}

	prep, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
	require.NoError(t, err)
	require.NoError(t, svc.MarkOrderPaidDev(ctx, prep.OrderID))

	token, err := svc.EnsureOrderQRToken(ctx, prep.OrderID)
	require.NoError(t, err)

	pub := keys.PublicKey()
	payload, err := qrsign.Verify(pub, token, time.Now())
	require.NoError(t, err)

	byProduct := map[string]int{}
	for _, l := range payload.Lines {
		byProduct[l.ProductID] = l.Quantity
	}
	// Bundle parent isn't redeemable; its two components are, at the bundle quantity.
	require.NotContains(t, byProduct, combo.ID, "bundle parent must not be a redeemable line")
	require.Equal(t, 2, byProduct[burger.ID])
	require.Equal(t, 2, byProduct[drink.ID])
	require.Len(t, payload.Lines, 2)
}

func TestQRConfig_PublicKeyEncodingRoundTrip(t *testing.T) {
	svc := NewQRKeySvc()
	pub := svc.PublicKey()

	served := base64.RawURLEncoding.EncodeToString(pub)
	decoded, err := decodePub(served)
	require.NoError(t, err)
	require.Equal(t, []byte(pub), []byte(decoded))
}
