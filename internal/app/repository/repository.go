package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Construction struct {
	ID                int
	ConstructionTitle string
	UseLife           string
	ImageKey          string
	VideoKey          string
	Description       string
}

type Application struct {
	ID                int
	ConstructionTitle string
	UseLife           string
	DateCorrection    string
	CuttingDate       string
	SamplesCount      int
	ImageKey          string
}

func (r *Repository) GetConstructions() ([]Construction, error) {
	constructions := []Construction{
		{
			ID:                1,
			ConstructionTitle: "Частокол",
			UseLife:           "20 лет",
			ImageKey:          "stockade.png",
			VideoKey:          "stockade.mp4",
			Description:       "Ограждающая конструкция из вертикально вкопанных заострённых брёвен, традиционно использовавшаяся для защиты поселений. Отличается простотой возведения и высокой механической прочностью. Типовой use-life: 20 лет",
		},
		{
			ID:                2,
			ConstructionTitle: "Опорные сваи",
			UseLife:           "40 лет",
			ImageKey:          "supportingPiles.png",
			VideoKey:          "supportingPiles.mp4",
			Description:       "Деревянные опорные элементы, забиваемые в грунт для передачи нагрузки от здания на более плотные слои почвы. Применяются в фундаментах на слабых грунтах и в условиях высокого уровня грунтовых вод. Типовой use-life: 40 лет",
		},
		{
			ID:                3,
			ConstructionTitle: "Деревянная кровля",
			UseLife:           "35 лет",
			ImageKey:          "woodenRoof.png",
			VideoKey:          "woodenRoof.mp4",
			Description:       "Традиционное кровельное покрытие из дранки или гонта, широко применявшееся в жилых и хозяйственных постройках. Отличается хорошей теплоизоляцией и устойчивостью к перепадам температуры, но требует правильной укладки и вентиляции. Типовой use-life: 35 лет",
		},
		{
			ID:                4,
			ConstructionTitle: "Сруб из бревна",
			UseLife:           "60 лет",
			ImageKey:          "logCabin.png",
			VideoKey:          "logCabin.mp4",
			Description:       "Несущая конструкция стен, собранная из горизонтально уложенных брёвен с угловыми врубками. Обеспечивает отличную теплоизоляцию и долговечность при правильной обработке древесины. Типовой use-life: 60 лет",
		},
		{
			ID:                5,
			ConstructionTitle: "Деревянная лестница",
			UseLife:           "35 лет",
			ImageKey:          "woodenLadder.png",
			VideoKey:          "woodenLadder.mp4",
			Description:       "Внутренняя или наружная конструкция для перемещения между этажами, изготовленная из массива дерева. Сочетает функциональность и эстетику, требует защитного покрытия для долговечности. Типовой use-life: 35 лет",
		},
		{
			ID:                6,
			ConstructionTitle: "Деревянная дверь",
			UseLife:           "50 лет",
			ImageKey:          "woodenDoor.png",
			VideoKey:          "woodenDoor.mp4",
			Description:       "Дверные полотна из массива древесины или клеёного бруса, применяемые как входные и межкомнатные. Обладают высокой звуко- и теплоизоляцией, экологичностью и долгим сроком службы. Типовой use-life: 50 лет",
		},
	}

	if len(constructions) == 0 {
		return nil, fmt.Errorf("список конструкций пуст")
	}

	return constructions, nil
}

func (r *Repository) GetConstruction(id int) (Construction, error) {
	constructions, err := r.GetConstructions()
	if err != nil {
		return Construction{}, err
	}

	for _, c := range constructions {
		if c.ID == id {
			return c, nil
		}
	}
	return Construction{}, fmt.Errorf("конструкция не найдена")
}

func (r *Repository) GetConstructionsByTitle(ConstructionTitle string) ([]Construction, error) {
	constructions, err := r.GetConstructions()
	if err != nil {
		return []Construction{}, err
	}

	var result []Construction
	for _, c := range constructions {
		if strings.Contains(strings.ToLower(c.ConstructionTitle), strings.ToLower(ConstructionTitle)) {
			result = append(result, c)
		}
	}

	return result, nil
}

func (r *Repository) GetApplications() ([]Application, error) {
	applications := []Application{
		{
			ID:                1,
			ConstructionTitle: "Деревянная кровля",
			UseLife:           "35 лет",
			DateCorrection:    "5 лет",
			CuttingDate:       "2020",
			SamplesCount:      1,
			ImageKey:          "woodenRoof.png",
		},
		{
			ID:                2,
			ConstructionTitle: "Сруб из бревна",
			UseLife:           "60 лет",
			DateCorrection:    "12 лет",
			CuttingDate:       "2013",
			SamplesCount:      1,
			ImageKey:          "logCabin.png",
		},
	}

	if len(applications) == 0 {
		return nil, fmt.Errorf("список заявок пуст")
	}

	return applications, nil
}
