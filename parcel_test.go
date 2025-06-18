package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

func initDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    );`)
	return err
}

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "parcel.db")
	require.NoError(t, err)
	err = initDB(db)
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	res, err := store.Get(number)
	assert.NoError(t, err)
	parcel.Number = number
	assert.Equal(t, res, parcel)

	err = store.Delete(number)
	assert.NoError(t, err)
	resD, err := store.Get(number)
	assert.Error(t, err)
	assert.Equal(t, Parcel{}, resD)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "parcel.db")
	require.NoError(t, err)
	err = initDB(db)
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	assert.NoError(t, err)

	res, err := store.Get(number)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, res.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "parcel.db")
	require.NoError(t, err)
	err = initDB(db)
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	statusList := []string{
		ParcelStatusRegistered,
		ParcelStatusSent,
		ParcelStatusDelivered,
	}

	for _, status := range statusList {
		err := store.SetStatus(number, status)
		assert.NoError(t, err)

		res, err := store.Get(number)
		assert.NoError(t, err)

		require.Equal(t, status, res.Status)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "parcel.db")
	require.NoError(t, err)
	err = initDB(db)
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		assert.NoError(t, err)
		assert.NotZero(t, id)

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	assert.NoError(t, err)

	assert.Equal(t, len(parcelMap), len(storedParcels))

	for _, parcel := range storedParcels {
		assert.NotEmpty(t, parcel.Address)
		assert.NotZero(t, parcel.Client)
		assert.NotEmpty(t, parcel.CreatedAt)
		assert.NotZero(t, parcel.Number)
		assert.NotEmpty(t, parcel.Status)

		parcel2, ok := parcelMap[parcel.Number]
		assert.True(t, ok)

		assert.Equal(t, parcel2, parcel)
	}
}
