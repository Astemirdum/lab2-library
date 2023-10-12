package circuit_breaker_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Astemirdum/library-service/pkg/circuit_breaker"
)

func Test_circuitBreaker_Call(t *testing.T) {
	type fields struct {
		recordLength     int
		timeout          time.Duration
		percentile       float64
		recoveryRequests int
	}
	type args struct {
		successfulService func() error
		failingService    func() error
	}

	successfulService := func() error {
		return nil
	}
	_ = successfulService

	failingService := func() error {
		return errors.New("service error")
	}
	_ = failingService

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successfulService",
			fields: fields{
				recordLength:     10,
				timeout:          2 * time.Second,
				percentile:       0.30,
				recoveryRequests: 100,
			},
			args: args{
				successfulService: successfulService,
				failingService:    failingService,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := circuit_breaker.NewCircuitBreaker(tt.fields.recordLength, tt.fields.timeout, tt.fields.percentile, tt.fields.recoveryRequests)
			for i := 0; i < 80; i++ {
				if err := cb.Call(tt.args.successfulService); err != nil {
					fmt.Printf("Service call failed: %s\n", err.Error())
				}
				fmt.Println(i, " ok ")
			}
			// Исполняем запросы с ошибкой
			fmt.Println("\nSending failing requests...")

			for i := 0; i < 40; i++ {
				if err := cb.Call(failingService); err != nil {
					fmt.Printf("%d Service call failed: %s\n", i, err.Error())
				}
			}

			// Ожидаем, чтобы CircuitBreaker перешел в half-open state
			fmt.Printf("\nWaiting for circuit breaker to switch to half-open state...\n")
			time.Sleep(3 * time.Second)

			// Исполняем запросы для перехода в closed state
			fmt.Println("Sending successful requests to recover...")
			for i := 0; i < 15; i++ {
				if err := cb.Call(successfulService); err != nil {
					fmt.Printf("Service call failed: %s\n", err.Error())
				}

				fmt.Printf("%d ok\n", i)
			}

			// Исполняем запросы с ошибкой для перехода обратно в open state
			fmt.Printf("\nSending failing requests to switch back to open state 1 ...\n\n")
			for i := 0; i < 40; i++ {
				if err := cb.Call(failingService); err != nil {
					fmt.Printf("%d Service call failed: %s\n", i, err.Error())
				}
			}

			// Ожидаем, чтобы CircuitBreaker перешел в half-open state
			fmt.Printf("\nWaiting for circuit breaker to switch to half-open state...\n")
			time.Sleep(3 * time.Second)

			// Исполняем запросы с ошибкой для перехода обратно в open state
			fmt.Printf("\nSending failing requests to switch back to open state 2 ...\n")
			for i := 0; i < 10; i++ {
				if err := cb.Call(failingService); err != nil {
					fmt.Printf("%d Service call failed: %s\n", i, err.Error())
				}
			}
		})
	}
}
