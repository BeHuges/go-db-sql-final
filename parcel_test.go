package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

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
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel

	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	assert.Equal(t, parcel, stored)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД

	err = store.Delete(parcel.Number)
	require.NoError(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавление новой посылки в БД, проверка отсутствия ошибки и наличия идентификатора

	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set address
	// обновление адреса, проверка отсутствия ошибки
	newAddress := "new test address"

	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	// получение добавленной посылки, проверка что адрес обновился

	stored, err := store.Get(parcel.Number)
	require.NoError(t, err)

	assert.Equal(t, newAddress, stored.Address)

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавление новой посылки в БД, проверка  отсутствия ошибки и наличия идентификатора

	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set status
	// обновление статуса, проверка отсутствия ошибки

	err = store.SetStatus(parcel.Number, ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получение добавленной посылки и проверка, что статус обновился
	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, stored.Status)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключение к БД
	require.NoError(t, err)
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
		id, err := store.Add(parcels[i]) // добавление новой посылки в БД, проверка  отсутствия ошибки и наличия идентификатора

		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) //  список посылок по идентификатору клиента, сохранённого в переменной client
	// проверка отсутствия ошибки
	// проверка, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// проверка, что все посылки из storedParcels есть в parcelMap
		// проверка, что значения полей полученных посылок заполнены верно

		_, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		assert.Equal(t, parcel, parcelMap[parcel.Number])
	}
}
