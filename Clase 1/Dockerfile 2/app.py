from fastapi import FastAPI
import os

app = FastAPI(title="Mi API Docker")

@app.get("/")
def read_root():
    return {"message": "Hola desde Docker!", "status": "success"}

@app.get("/health")
def health_check():
    return {"status": "healthy","port":os.getenv("PORT", "8000")}

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0",port=port)