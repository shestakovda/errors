package errx

import "errors"

type Error interface {
	/*
		Error - стандартный интерфейс ошибки.

		* Возвращает исходный текст, переданный в конструктор
		* Если был вызов WithArgs, возвращает форматированный исходный текст
	*/
	error

	/*
		Unwrap - движение по цепочке вниз, получение следующей ошибки

		* Перегрузка логики сравнения для использования со стандартным errors.Unwrap
		* Если цепочка кончилась, возвращает nil
	*/
	Unwrap() error

	/*
		Is - сравнение с ошибками в цепочке.

		* Перегрузка логики сравнения для использования со стандартным errors.Is
		* Если передан nil, возвращает false
		* Сначала сравнение на прямое равенство, затем по сообщению Error()
		* После - аналогичное сравнение с прототипом (если он быть создан с помощью With*)
		* Если был WithReason, то сравнение идет дальше по цепочке
	*/
	Is(err error) bool

	/*
		WithStack - формирование стека вызовов для локализации ошибки.

		* Стек только в текущей горутине выполнения
		* Для избежания проблем с гонками при обновлении, делает копию шаблонной ошибки
		* Если стек уже собран и копия выполнена, повторный вызов не меняет содержимого
		* Напрямую вызывается только если не нужны ни причина, ни форматирование
	*/
	WithStack() Error

	/*
		WithReason - добавление исходной ошибки для понимания причин возникновения (аналог xerrors.Wrap).

		* Добавляет указанную ошибку в цепочку
		* Автоматически вызывает WithStack
	*/
	WithReason(err error) Error

	/*
		WithDetail - добавление аргументов для форматирования сообщения.

		* Устанавливает форматированное детальное описание ошибки для пользователя
		* Автоматически вызывает WithStack
	*/
	WithDetail(tpl string, args ...interface{}) Error

	/*
		WithDebug - добавление объекта отладочных данных.

		* Автоматически вызывает WithStack
	*/
	WithDebug(dbg Debug) Error

	/*
		Export - конвертация в нейтральное от реализации представление
	*/
	Export() *View
}

type Debug map[string]interface{}

func As(err error, target interface{}) bool { return errors.As(err, target) }
func Is(err error, targets ...error) bool {
	for i := range targets {
		if errors.Is(err, targets[i]) {
			return true
		}
	}
	return false
}
func New(text string) Error  { return newErrorV1(text) }
func Unwrap(err error) error { return errors.Unwrap(err) }

// View - представление ошибки для простой работы с содержимым
type View struct {
	Next   *View
	Text   string
	Detail string
	Stack  []string
	Debug  map[string]string
}
