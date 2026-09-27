package bch

// fieldTables хранит таблицы степеней и
// логарифмов поля GF(2^7). Таблица степеней
// удвоена, чтобы сумма двух логарифмов
// не требовала взятия остатка.
type fieldTables struct {
	exp [SIZE_GROUP << 1]uint8
	log [SIZE_FIELD]uint8
}

// TABLES — таблицы поля, строятся один раз.
var TABLES = buildTables()

// GENERATOR — порождающий полином кода,
// бит с номером d — коэффициент при x^d.
var GENERATOR = buildGenerator()

// buildTables строит таблицы степеней
// примитивного элемента и логарифмов.
//
// Возвращает: заполненные таблицы поля.
func buildTables() fieldTables {
	var tables fieldTables
	value := 1
	for indPow := 0; indPow < SIZE_GROUP; indPow++ {
		tables.exp[indPow] = uint8(value)
		tables.exp[indPow+SIZE_GROUP] = uint8(value)
		tables.log[value] = uint8(indPow)
		value <<= 1
		if value&MASK_FIELD_TOP != 0 {
			value ^= POLY_FIELD
		}
	}
	return tables
}

// mulField умножает два элемента поля.
//
// Параметры:
//   - left, right: элементы поля, 0..127.
//
// Возвращает: произведение, 0..127.
func mulField(left, right uint8) uint8 {
	product := uint8(0)
	if left != 0 && right != 0 {
		sumLog := int(TABLES.log[left]) +
			int(TABLES.log[right])
		product = TABLES.exp[sumLog]
	}
	return product
}

// divField делит элемент поля на ненулевой.
//
// Параметры:
//   - numer: делимое, 0..127.
//   - denom: делитель, 1..127.
//
// Возвращает: частное, 0..127.
func divField(numer, denom uint8) uint8 {
	quotient := uint8(0)
	if numer != 0 {
		diffLog := int(TABLES.log[numer]) +
			SIZE_GROUP - int(TABLES.log[denom])
		quotient = TABLES.exp[diffLog]
	}
	return quotient
}

// powAlpha возводит примитивный элемент
// в неотрицательную степень.
//
// Параметры:
//   - exponent: показатель, >= 0.
//
// Возвращает: элемент поля, 1..127.
func powAlpha(exponent int) uint8 {
	return TABLES.exp[exponent%SIZE_GROUP]
}

// minimalPoly строит минимальный полином
// элемента alpha^indRoot над GF(2) и
// отмечает его циклотомический класс.
//
// Параметры:
//   - indRoot: показатель корня, 1..126.
//   - isUsed: отметки использованных
//     показателей; дополняется классом.
//
// Возвращает: биты полинома, бит d — x^d.
func minimalPoly(
	indRoot int,
	isUsed *[SIZE_GROUP]bool,
) uint64 {
	var coeffs [SIZE_DATA + 1]uint8
	coeffs[0] = 1
	degree := 0
	for power := indRoot; !isUsed[power]; {
		isUsed[power] = true
		root := powAlpha(power)
		for indCoef := degree + 1; indCoef > 0; indCoef-- {
			coeffs[indCoef] = coeffs[indCoef-1] ^
				mulField(coeffs[indCoef], root)
		}
		coeffs[0] = mulField(coeffs[0], root)
		degree++
		power = (power << 1) % SIZE_GROUP
	}
	poly := uint64(0)
	for indCoef := 0; indCoef <= degree; indCoef++ {
		poly |= uint64(coeffs[indCoef]) << indCoef
	}
	return poly
}

// mulBinary перемножает полиномы над GF(2).
//
// Параметры:
//   - left, right: биты полиномов; сумма
//     степеней не больше 63.
//
// Возвращает: биты произведения.
func mulBinary(left, right uint64) uint64 {
	product := uint64(0)
	for rest := right; rest != 0; rest >>= 1 {
		if rest&1 != 0 {
			product ^= left
		}
		left <<= 1
	}
	return product
}

// buildGenerator строит порождающий полином
// как произведение различных минимальных
// полиномов корней alpha^1..alpha^20.
//
// Возвращает: биты полинома степени 63.
func buildGenerator() uint64 {
	var isUsed [SIZE_GROUP]bool
	generator := uint64(1)
	for indRoot := 1; indRoot <= CNT_SYNDROMES; indRoot++ {
		if !isUsed[indRoot] {
			generator = mulBinary(
				generator,
				minimalPoly(indRoot, &isUsed),
			)
		}
	}
	return generator
}
