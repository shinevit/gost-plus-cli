package option

import "fmt"

// ===== Flow DSL ====
type Flow[A, R any] struct {
	fallback R
	run      func(Option[A]) Option[R]
}

func (f Flow[A, R]) Run(input A) (err error, result R) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[Flow] Recovered from panic: %v", r)
			fmt.Println(err)
			result = f.fallback
		}
	}()

	result = f.run(Some(&input)).GetOrElse(f.fallback)
	return nil, result
}

// Chain
// *A -> *B
func Chain[A, B any](f func(*A) *B) func(Option[*A]) Option[*B] {
	return func(opt Option[*A]) Option[*B] {
		if opt.value == nil || *opt.value == nil {
			return None[*B]()
		}
		result := f(*opt.value)
		return Some(&result)
	}
}

// *A -> B
func MapResult[A, B any](f func(*A) B) func(Option[*A]) Option[B] {
	return func(opt Option[*A]) Option[B] {
		if opt.value == nil || *opt.value == nil {
			return None[B]()
		}
		result := f(*opt.value)
		return Some(&result)
	}
}

// Fallback function for Flow
func GetOrElse[T any](fallback T) func(Option[T]) T {
	return func(opt Option[T]) T {
		return opt.GetOrElse(fallback)
	}
}

// Flow helpers methods
func Flow1[A, B, C any](
	mapFn func(Option[A]) Option[B],
	defaultFn func(Option[B]) C,
) Flow[A, C] {
	return Flow[A, C]{
		fallback: defaultFn(None[B]()),

		run: func(opt Option[A]) Option[C] {
			if mapFn == nil || defaultFn == nil {
				panic("Flow1: one or more steps are nil")
			}
			result := defaultFn(mapFn(opt))
			return Some(&result)
		},
	}
}

func Flow2[A, B, C, D any](
	step1 func(Option[A]) Option[B],
	mapFn func(Option[B]) Option[C],
	defaultFn func(Option[C]) D,
) Flow[A, D] {
	return Flow[A, D]{
		fallback: defaultFn(None[C]()),

		run: func(opt Option[A]) Option[D] {
			if step1 == nil || mapFn == nil || defaultFn == nil {
				panic("Flow2: one or more steps are nil")
			}
			result := defaultFn(mapFn(step1(opt)))
			return Some(&result)
		},
	}
}

func Flow3[A, B, C, D, E any](
	step1 func(Option[A]) Option[B],
	step2 func(Option[B]) Option[C],
	mapFn func(Option[C]) Option[D],
	defaultFn func(Option[D]) E,
) Flow[A, E] {
	return Flow[A, E]{
		fallback: defaultFn(None[D]()),

		run: func(opt Option[A]) Option[E] {
			if step1 == nil || step2 == nil || mapFn == nil {
				panic("Flow3: one or more steps are nil")
			}
			result := defaultFn(mapFn(step2(step1(opt))))
			return Some(&result)
		},
	}
}

func Flow4[A, B, C, D, E, F any](
	step1 func(Option[A]) Option[B],
	step2 func(Option[B]) Option[C],
	step3 func(Option[C]) Option[D],
	mapFn func(Option[D]) Option[E],
	defaultFn func(Option[E]) F,
) Flow[A, F] {
	return Flow[A, F]{
		fallback: defaultFn(None[E]()),

		run: func(opt Option[A]) Option[F] {
			if step1 == nil || step2 == nil || step3 == nil || mapFn == nil || defaultFn == nil {
				panic("Flow4: one or more steps are nil")
			}
			result := defaultFn(mapFn(step3(step2(step1(opt)))))
			return Some(&result)
		},
	}
}

func Flow5[A, B, C, D, E, F, G any](
	step1 func(Option[A]) Option[B],
	step2 func(Option[B]) Option[C],
	step3 func(Option[C]) Option[D],
	step4 func(Option[D]) Option[E],
	mapFn func(Option[E]) Option[F],
	defaultFn func(Option[F]) G,
) Flow[A, G] {
	return Flow[A, G]{
		fallback: defaultFn(None[F]()),

		run: func(opt Option[A]) Option[G] {
			if step1 == nil || step2 == nil || step3 == nil ||
				step4 == nil || mapFn == nil || defaultFn == nil {
				panic("Flow5: one or more steps are nil")
			}

			result := defaultFn(mapFn(step4(step3(step2(step1(opt))))))
			return Some(&result)
		},
	}
}
