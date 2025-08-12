ARG BASE_IMAGE=pytorch/pytorch:2.7.0-cuda12.8-cudnn9-devel
FROM ${BASE_IMAGE}

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -yq git build-essential

WORKDIR /home/comfy
ARG COMFYUI_URL=https://github.com/comfyanonymous/ComfyUI
ARG COMFYUI_COMMIT=e18f53cca9062cc6b165e16712772437b80333f2
ARG COMFYUI_MANAGER_URL=https://github.com/ltdrdata/ComfyUI-Manager
ARG COMFYUI_MANAGER_COMMIT=5978b6c9ee82e46bcd78bf017d79923fa7888489
RUN git clone ${COMFYUI_URL} ComfyUI && \
    cd ComfyUI && \
    git checkout ${COMFYUI_COMMIT} && \
    python -m pip install -r requirements.txt && \
    cd custom_nodes && \
    git clone ${COMFYUI_MANAGER_URL} ComfyUI-Manager && \
    cd ComfyUI-Manager && \
    git checkout ${COMFYUI_MANAGER_COMMIT} && \
    python -m pip install -r requirements.txt

# Add Tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

# Install PyTorch
ARG EXTRA_TORCH_INDEX_URL=https://download.pytorch.org/whl/cu128
RUN pip install --no-cache-dir torch torchvision torchaudio --extra-index-url "${EXTRA_TORCH_INDEX_URL}"

ENV COMFYUI_PATH=/home/comfy/ComfyUI
WORKDIR ${COMFYUI_PATH}

EXPOSE 8188
ENTRYPOINT ["/tini", "--"]
CMD ["python", "main.py", "--listen", "--port", "8188"]
