# k3s para pruebas locales
> Página oficial: https://k3s.io/

k3s es una distribución ligera y optimizada específicamente para kubernetes la cual fue creada por Rancher (ahora parte de SUSE) y alojada en la CNCF.


## Instalación

Primero actualizamos el sistema:
```bash
sudo apt update && sudo apt upgrade -y
```

Luego instalamos el script oficial:
```bash
curl -sfL https://get.k3s.io | sh -
```

Y como tal ya estaría instalado, para verificar la instalación usamos:
```bash
sudo systemctl status k3s
```

Si llegamos hasta este punto tendremos que usar `sudo` en todo momento, para no hacerlo ejecutamos lo siguiente:
```bash
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
```

## Ejemplo básico
```bash
kubectl create deployment nginx --image=nginx
kubectl expose deployment nginx --port=80 --type=NodePort
kubectl get svc
```