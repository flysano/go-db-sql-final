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

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	res, err := store.Get(number)
	assert.NoError(t, err)
	parcel.Number = number
	assert.Equal(t, res, parcel)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
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
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	assert.NoError(t, err)

	res, err := store.Get(number)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, res.Address)
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
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

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
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

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
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
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		assert.NoError(t, err)
		assert.NotZero(t, id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	assert.NoError(t, err)
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(parcelMap), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		assert.NotEmpty(t, parcel.Address)
		assert.NotZero(t, parcel.Client)
		assert.NotEmpty(t, parcel.CreatedAt)
		assert.NotZero(t, parcel.Number)
		assert.NotEmpty(t, parcel.Status)
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		parcel2, ok := parcelMap[parcel.Number]
		assert.True(t, ok)

		assert.Equal(t, parcel2, parcel)
	}
}
