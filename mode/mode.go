//============================================//
// Пакет mode                                 //
//                                            //
// Назначение:                                //
//   таблица режимов сокрытия данных:         //
//   aggressive   — для опытов с YouTube,     //
//   experimental — промежуточный,            //
//   comfortable  — для локального хранения.  //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package mode

// Params — параметры одного режима.
type Params struct {
	Id           uint8
	Name         string
	Mix          float64
	CntCopies    int
	IsBitShuffle bool
	Crf          string
}

// MODES — таблица режимов.
var MODES = []Params{
	{
		Id:           ID_AGGRESSIVE,
		Name:         NAME_AGGRESSIVE,
		Mix:          MIX_AGGRESSIVE,
		CntCopies:    CNT_COPIES_AGGRESSIVE,
		IsBitShuffle: true,
		Crf:          CRF_AGGRESSIVE,
	},
	{
		Id:           ID_EXPERIMENTAL,
		Name:         NAME_EXPERIMENTAL,
		Mix:          MIX_EXPERIMENTAL,
		CntCopies:    CNT_COPIES_EXPERIMENTAL,
		IsBitShuffle: true,
		Crf:          CRF_EXPERIMENTAL,
	},
	{
		Id:           ID_COMFORTABLE,
		Name:         NAME_COMFORTABLE,
		Mix:          MIX_COMFORTABLE,
		CntCopies:    CNT_COPIES_COMFORTABLE,
		IsBitShuffle: false,
		Crf:          CRF_COMFORTABLE,
	},
}

// find ищет режим по предикату.
//
// Параметры:
//   - isWanted: предикат искомого режима.
//
// Возвращает: параметры режима и признак
// того, что режим найден.
func find(
	isWanted func(Params) bool,
) (Params, bool) {
	var found Params
	isFound := false
	cntModes := len(MODES)
	for indMode := 0; !isFound && indMode < cntModes; indMode++ {
		found = MODES[indMode]
		isFound = isWanted(found)
	}
	return found, isFound
}

// ByName ищет режим по имени.
//
// Параметры:
//   - name: имя режима.
//
// Возвращает: параметры режима и признак
// того, что режим найден.
func ByName(name string) (Params, bool) {
	return find(func(params Params) bool {
		return params.Name == name
	})
}

// ById ищет режим по идентификатору.
//
// Параметры:
//   - idMode: идентификатор режима.
//
// Возвращает: параметры режима и признак
// того, что режим найден.
func ById(idMode uint8) (Params, bool) {
	return find(func(params Params) bool {
		return params.Id == idMode
	})
}
