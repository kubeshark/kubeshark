package acceptanceTests

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
	"os/exec"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmd := exec.Command(cliPath, tapCmdArgs...)
	t.Logf("running command: %v", tapCmd.String())

	t.Cleanup(func() {
		if err := CleanupCommand(tapCmd); err != nil {
			t.Logf("failed to cleanup tap command, err: %v", err)
		}
	})

	if err := tapCmd.Start(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	ctx := context.Background()

	kubernetesProvider, err := NewKubernetesProvider()
	if err != nil {
		t.Errorf("failed to create k8s provider, err %v", err)
		return
	}

	redisExternalIp, err := kubernetesProvider.GetServiceExternalIp(ctx, DefaultNamespaceName, "redis")
	if err != nil {
		t.Errorf("failed to get redis external ip, err: %v", err)
		return
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%v:6379", redisExternalIp),
	})

	for i := 0; i < DefaultEntriesCount/5; i++ {
		requestErr := rdb.Ping(ctx).Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < DefaultEntriesCount/5; i++ {
		requestErr := rdb.Set(ctx, "key", "value", -1).Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < DefaultEntriesCount/5; i++ {
		requestErr := rdb.Exists(ctx, "key").Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < DefaultEntriesCount/5; i++ {
		requestErr := rdb.Get(ctx, "key").Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < DefaultEntriesCount/5; i++ {
		requestErr := rdb.Del(ctx, "key").Err()
		if requestErr != nil {
			t.Errorf("failed to send redis request, err: %v", requestErr)
			return
		}
	}

	RunCypressTests(t, "npx cypress@9.5.4 run --spec  \"cypress/integration/tests/Redis.js\"")
}

func TestAmqp(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmd := exec.Command(cliPath, tapCmdArgs...)
	t.Logf("running command: %v", tapCmd.String())

	t.Cleanup(func() {
		if err := CleanupCommand(tapCmd); err != nil {
			t.Logf("failed to cleanup tap command, err: %v", err)
		}
	})

	if err := tapCmd.Start(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	ctx := context.Background()

	kubernetesProvider, err := NewKubernetesProvider()
	if err != nil {
		t.Errorf("failed to create k8s provider, err %v", err)
		return
	}

	rabbitmqExternalIp, err := kubernetesProvider.GetServiceExternalIp(ctx, DefaultNamespaceName, "rabbitmq")
	if err != nil {
		t.Errorf("failed to get RabbitMQ external ip, err: %v", err)
		return
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%v:5672/", rabbitmqExternalIp))
	if err != nil {
		t.Errorf("failed to connect to RabbitMQ, err: %v", err)
		return
	}
	defer conn.Close()

	// Temporary fix for missing amqp entries
	time.Sleep(10 * time.Second)

	for i := 0; i < DefaultEntriesCount/5; i++ {
		ch, err := conn.Channel()
		if err != nil {
			t.Errorf("failed to open a channel, err: %v", err)
			return
		}

		exchangeName := "exchange"
		err = ch.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil)
		if err != nil {
			t.Errorf("failed to declare an exchange, err: %v", err)
			return
		}

		q, err := ch.QueueDeclare("queue", true, false, false, false, nil)
		if err != nil {
			t.Errorf("failed to declare a queue, err: %v", err)
			return
		}

		routingKey := "routing_key"
		err = ch.QueueBind(q.Name, routingKey, exchangeName, false, nil)
		if err != nil {
			t.Errorf("failed to bind the queue, err: %v", err)
			return
		}

		err = ch.Publish(exchangeName, routingKey, false, false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte("message"),
			})
		if err != nil {
			t.Errorf("failed to publish a message, err: %v", err)
			return
		}

		msgChan, err := ch.Consume(q.Name, "Consumer", true, false, false, false, nil)
		if err != nil {
			t.Errorf("failed to create a consumer, err: %v", err)
			return
		}

		select {
		case <-msgChan:
			break
		case <-time.After(3 * time.Second):
			t.Errorf("failed to consume a message on time")
			return
		}

		err = ch.ExchangeDelete(exchangeName, false, false)
		if err != nil {
			t.Errorf("failed to delete the exchange, err: %v", err)
			return
		}

		_, err = ch.QueueDelete(q.Name, false, false, false)
		if err != nil {
			t.Errorf("failed to delete the queue, err: %v", err)
			return
		}

		ch.Close()
	}

	RunCypressTests(t, "npx cypress@9.5.4 run --spec  \"cypress/integration/tests/Rabbit.js\"")
}
