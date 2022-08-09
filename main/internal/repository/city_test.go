package repository

import (
	"fmt"
	"testing"
)

func Test_insert(t *testing.T) {
	for idx, name := range []string{
		"Футбол",
		"Пиво",
		"Зависать с друзьями",
		"Книги",
		"Рыбалка",
		"Мода",
		"Красота",
		"Фитнес",
		"Правильное питание",
		"Автомобили",
		"Сериалы",
		"Кино",
		"Музыка",
		"Изобразительное искусство",
		"Наука",
		"Семья",
		"Отношения",
		"Путешествия",
		"Домашние животные",
	} {
		fmt.Println(fmt.Sprintf("(%d, '%s'),", idx+1, name))
	}
}