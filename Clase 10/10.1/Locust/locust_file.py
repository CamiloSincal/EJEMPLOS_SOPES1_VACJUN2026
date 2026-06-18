from locust import HttpUser, task, between
class UsuarioBasico(HttpUser):

    wait_time = between(1, 3)

    @task(3) 

    def ver_home(self):
        self.client.get("/")

    @task(1)
    def ver_messages(self):
        self.client.get("/messages")