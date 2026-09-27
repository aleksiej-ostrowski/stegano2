package merge

// Имена временных файлов
const PREFIX_TEMP = ".stegano2-part-"

// Константы хеш-функции SplitMix64 для
// заполнения свободных ячеек
const (
	MIX_DUST_GAMMA  = 0x9E3779B97F4A7C15
	MIX_DUST_FIRST  = 0xBF58476D1CE4E5B9
	MIX_DUST_SECOND = 0x94D049BB133111EB
	SHIFT_DUST_HIGH = 30
	SHIFT_DUST_MID  = 27
	SHIFT_DUST_LOW  = 31
)
