# RabbitMQ y Dapr en k8s

---

## RabbitMQ
RabitMQ es un *broker de mensajería*, es decir, una entidad o software que funciona como un intermediario entre sistemas o aplicaciones para hacer más facil la comunicación e intercambio de mensajería. La idea es implementarlo en arquitecturas de mensajería para evitar comunicación directa entre sistemas y suele implementarse por sus características de desacoplamiento de aplicaciones, comunicación asíncrona (sin bloqueos) y la garantía de entrega de datos.

![diag-rabbitmq](./imgs/rabbitmq.PNG)

De manera general RabbitMQ se compone de:
- **Productor (Producer):** Aplicacion o servicio que crea y envía mensajes al sistema.
- **Intercambiador (Exchange):** Quien recibe los mensajes de los productores y determina a qué cola deben dirigirse utilizando reglas de enrutamiento. Por lo general estas reglas son:
    - *Direct:* Un intercambiador de tipo direct enrutará el mensaje a la cola que tenga una llave de enlace igual a la llave de enrutamiento que el mensaje.
    - *Fanout:* El intercambiador fanout envía los mensajes hacia todas las colas que tenga enlazada a si mismo.
    - *Topic:* Un intercambiador de tipo topic implementa *wildcard match* para buscar patrones entre la llave de enrutamiento y el patrón de enrutamiento especificado en el enlace o binding.
    - *Headers:* Este tipo de intercambiador usará la información del encabezado del mensaje para enrutarlo a una cola.
- **Colas (Queues):** Lugar donde se acumulan los mensajes hasta que son procesados por un consumidor.
- **Enlace (Binding):** Reglas de conexión entre el intercambiador y la cola que indican al exchange qué mensajes específicos pueden entrar a una cola determinada.- **Consumidor (Consumer):** La aplicación o servicio que se conecta a una cola y recibe los mensajes para procesarlos.

## Dapr
Dapr por otro lado corresponde a un software intermedio (middleware) que se utiliza para realizar la comunicación o intercambio de información entre diferentes sistemas. A diferencia de RabbitMQ Dapr no ofrece directamente la infraestructura para esa comunciación, únicamente utiliza una arquitectura de *sidecar* (un proceso separado que se ejecuta junto a alguna aplicación). Nuestro código se comunicará con este sidecar a través de APIs estándar (HTTP o gRPC), evitando que tengamos que escribir código específico los brokers de mensajería u otros.

![diag-rabbitmq](./imgs/dapr-diagram.png)
> Imagen obtenida de la documentaicón de dapr (https://docs.dapr.io/developing-applications/building-blocks/pubsub/howto-publish-subscribe/)

---

## Configuración de RabbitMQ en Kubernetes

La forma más rápida para desplegar RabbitMQ  de en un clúster de Kubernetes es utilizar el *RabbitMQ Cluster Operator*.

Iniciamos creando un namespace por organización e instalamos el operador (si se quiere no es necesario implementar namespaces):

```bash
kubectl create namespace rabbitmq-system

kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml
```

Lo siguiente es hacer crear un YAML de rabbitmq y un apply de este, para este ejemplo se utilizará el siguiente:
```yml
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:dapr + RabbitMQ
  name: rabbitmq-cluster
  namespace: rabbitmq-system
spec:
  replicas: 1
  image: rabbitmq:3.13-management
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 250m
      memory: 512Mi
  persistence:
    storageClassName: standard
    storage: 10Gi
  service:
    type: ClusterIP
```

Hacemos el apply:
```bash
kubectl apply -f rabbitmq.yml
```

Y podemos verificar todo con:
```bash
kubectl get rabbitmqcluster -n rabbitmq-system
kubectl get pods -n rabbitmq-system -w
```

Podemos obtener las credenciales de rabbitMQ con:
```bash
kubectl get secret rabbitmq-cluster-default-user -n rabbitmq-system -o jsonpath='{.data.username}' | base64 -d
kubectl get secret rabbitmq-cluster-default-user -n rabbitmq-system -o jsonpath='{.data.password}' | base64 -d
```
## EJ1 - Producer + RabbitMQ + Consumer
La primera forma de comunicar servicios es codificando directamente el producto y consumidor. Para este caso se proporcionan el código de ambas en la carpeta `Scripts` dentro de `Manifiestos`.

Dentro de cada carpeta dentro de `Scripts` encontraremos tanto el código fuente como el archivo de Docker para crear la imagen del servicio. 
> Dentro del código hay comentarios para entender que hace cada fragmento, se hace de esta manera para no saturar este README

Para reproducir este ejemplo lo primero es crear las imagenes de Docker del `consumer` y `publisher` y cargarlas a nuestro servidor zot:

```bash
# Dentro de la carpeta Consumer
docker build -t <IP-VM-ZOT>:5000/consumer:v1 .

# Dentro de la carpeta Publisher
docker build -t <IP-VM-ZOT>:5000/publisher:v1 .

# Cargamos en zot
docker push <IP-VM-ZOT>:5000/consumer:v1
docker push <IP-VM-ZOT>:5000/publisher:v1
```

Con las imagenes creadas en zot ya solo nos queda hacer un apply de nuestro deployment y nuestro service, tanto para el Consumer como para el publisher.
> Para este ejemplo no se profundizará en los archivos yml, ya que son conceptos que se tocan en clases anteriores

Lo más imporante de ambos deployments es:
```yml
# Es necesario cambiar esta linea para el consumidor y el publisher
image: <DIRECCION-ZOT>/<IMAGEN>
```

```yml
# 1.Es necesario cambiar guest:guest por las credenciales de RabbitMQ
# 2. El valor de RABBITMQ_QUEUE debe ser igual tanto para el consumer como para el publisher 
env:
        - name: RABBITMQ_URL
          value: "amqp://guest:guest@rabbitmq-cluster.rabbitmq-system.svc.cluster.local:5672/"
        - name: RABBITMQ_QUEUE
          value: "mensajes"
        readinessProbe:
```

De los manifiestos de servicios lo único a tener en cuenta es que el nombre de la app del selector correspondan a los nombres de los deployments respectivos.
```yml
# Fragmento del .yml del service del consumer
spec:
  selector:
    app: rabbitmq-consumer # Esto debe ser igual al deployment a aplicarle el servicio
```

Con los archivos configurados y estando conectados al clúster simplemente ejecutamos los siguiente:
```bash
# Para el consumer
kubectl apply -f consumer-deployment.yml
kubectl apply -f consumer-service.yml

# Para el publisher
kubectl apply -f publisher-deployment.yml
kubectl apply -f publisher-service.yml
```

Para ver los logs de esta configuración hacemos lo siguiente:
```bash
# Para el consumer
kubectl logs deployment/rabbitmq-consumer

# Para el publisher
kubectl logs deployment/rabbitmq-publisher
```

## EJ2 - dapr + RabbitMQ

## Configuración inicial de Dapr
Como se mencionó anteriormente, dapr es posible de implementar sobre diferentes brokers de mensajería. Para poder implementarlo sobre RabbitMQ se deben seguir los siguientes pasos.

Iniciaremos instalando el CLI de Dapr en el clúster de k8s:
```bash
# Instalación
wget -q https://raw.githubusercontent.com/dapr/cli/master/install/install.sh -O - | /bin/bash

# Para verificar la instalación
dapr -h
```

Con el CLI instalado podemos instalar directamente Dapr en Kubernetes:
```bash
# Instalación
dapr init --kubernetes

# Verificación
dapr status -k
```

Podemos acceder a un dashboard con:
```bash
# Para dapr 1.13+ es necesario instalar el dasboard con:
# Instalar el dashboard (ajusta la versión según tu Dapr)
helm repo add dapr https://dapr.github.io/helm-charts/
helm repo update

helm install dapr-dashboard dapr/dapr-dashboard \
  --namespace dapr-system

# Acceder al dashboard
kubectl port-forward svc/dapr-dashboard -n dapr-system 8080:8080 --address 0.0.0.0
```

Con el port-forward ejecutad, y si se está usando GCP, damos click en la opción de la siguiente imagen para poder visualizar el dasboard:
![cloudshell](./imgs/cloudshell.PNG)

Para el siguiente paso tendremos que obtener o tener a la mano las credenciales de RabbitMQ como hicimos anteriormente. Con esto en mente haremos un apply del archivo `Dapr.yml` que está en la carpeta manifiestos:
>Es necesario cambiar las credenciales de RabbitMQ en el archivo yml tanto en `connectionString` como en `username` y `password`

```bash
kubectl apply -f Dapr.yml
```
### Publisher y Subscriber
Tanto para el publisher como para subscriber usaremos imagenes de archivos de go, estos corresponden a `dapr_publisher.go` que se encuentra dentro de la subcarpeta publisher de la carpeta Dapr y `dapr_subscriber.go` dentro de la subcarpeta subscriber de la carpeta Dapr. Estando dentro de la carpeta correspondiente a cada script construimos y hacemos push de las imagenes:

```bash
# publisher
docker build -t <DIRECCION-ZOT>/publisher:latest .
docker push <DIRECCION-ZOT>/publisher:latest

# subscriber
docker build -t <DIRECCION-ZOT>/subscriber:latest .
docker push <DIRECCION-ZOT>/subscriber:latest
```

Con las imagenes en nuestro zot solo modificamos los deployments de Dapr, tanto para el publisher como para el subscriber, de tal manera que la imagen apunte al zot:
```bash
# publisher
kubectl apply -f publisher.yml

# subscriber
kubectl apply -f subscriber.yml
```

Finalmente podemos ver los logs del publisher y subscriber con:
```bash
# publisher
kubectl logs deployment/publisher -c dapr-go-publisher

# subscriber
kubectl logs deployment/subscriber -c dapr-go-subscriber
```

Si queremos ver los logs del sidecar de dapr usamos:
```bash
# publisher
kubectl logs deployment/publisher -c daprd
# subscriber
kubectl logs deployment/subscriber -c daprd
```

## Implementación de PubSub de Dapr
Para poder ver la interacción o logs en tiempo real:
```bash
#open logs for publisher
kubectl logs deploy/publisher --all-containers=true -f
#open logs for subscriber
kubectl logs deploy/subscriber --all-containers=true -f
```

Para exponer la API del publisher para enviar datos al pubsub que usa RabbitMQ:
```bash
kubectl port-forward svc/publisher 4500:4500 --address 0.0.0.0
```

En linux podemos hacer un curlo o en windows usar postman:
```bash
curl -X POST http://localhost:4500/publish \
-H "Content-Type: application/json" \
-d '{"data": "data1"}'
```