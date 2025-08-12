package comfyui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sko00o/comfyui-go/ws/message"
)

func TestMessageObj_UnmarshalJSON(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		o       MessageObj
		args    args
		wantErr bool
	}{
		{
			name: "execution_start",
			o: MessageObj{
				Type: message.ExecutionStart,
				Data: &message.DataExecution{
					ExecutionBase: message.ExecutionBase{
						PromptID: "4c1484b8-8b71-4c92-a1e8-6179c56fe67c",
					},
					Timestamp: func() *int64 { v := int64(1722433626779); return &v }(),
					Nodes:     nil,
				},
			},
			args: args{
				p: []byte(`[
  "execution_start",
  {
    "prompt_id": "4c1484b8-8b71-4c92-a1e8-6179c56fe67c",
    "timestamp": 1722433626779
  }
]`),
			},
			wantErr: false,
		},
		{
			name: "execution_cached",
			o: MessageObj{
				Type: message.ExecutionCached,
				Data: &message.DataExecution{
					ExecutionBase: message.ExecutionBase{
						PromptID: "4c1484b8-8b71-4c92-a1e8-6179c56fe67c",
					},
					Timestamp: func() *int64 { v := int64(1722433626780); return &v }(),
					Nodes: []string{
						"9",
						"8",
						"6",
						"4",
						"5",
						"1",
						"7",
					},
				},
			},
			args: args{
				p: []byte(`[
  "execution_cached",
  {
    "nodes": [
      "9",
      "8",
      "6",
      "4",
      "5",
      "1",
      "7"
    ],
    "prompt_id": "4c1484b8-8b71-4c92-a1e8-6179c56fe67c",
    "timestamp": 1722433626780
  }
]`),
			},
			wantErr: false,
		},
		{
			name: "execution_success",
			o: MessageObj{
				Type: message.ExecutionSuccess,
				Data: &message.DataExecution{
					ExecutionBase: message.ExecutionBase{
						PromptID: "4c1484b8-8b71-4c92-a1e8-6179c56fe67c",
					},
					Timestamp: func() *int64 { v := int64(1722433626780); return &v }(),
					Nodes:     nil,
				},
			},
			args: args{
				p: []byte(`[
  "execution_success",
  {
    "prompt_id": "4c1484b8-8b71-4c92-a1e8-6179c56fe67c",
    "timestamp": 1722433626780
  }
]`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.o

			var got MessageObj
			if err := json.Unmarshal(tt.args.p, &got); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.EqualValues(t, want, got)
		})
	}
}

func TestPromptObj_UnmarshalJSON(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		o       PromptObj
		args    args
		wantErr bool
	}{
		{
			name: "001",
			o: PromptObj{
				Num:       92,
				PromptID:  "4c1484b8-8b71-4c92-a1e8-6179c56fe67c",
				Workflow:  []byte(`{"1":{"inputs":{"ckpt_name":"v1-5-pruned-emaonly.safetensors"},"class_type":"CheckpointLoaderSimple","_meta":{"title":"Load Checkpoint"}},"4":{"inputs":{"text":"beautiful scenery nature glass bottle landscape, purple galaxy bottle,","clip":["1",1]},"class_type":"CLIPTextEncode","_meta":{"title":"CLIP Text Encode (Prompt)"}},"5":{"inputs":{"text":"text, watermark","clip":["1",1]},"class_type":"CLIPTextEncode","_meta":{"title":"CLIP Text Encode (Prompt)"}},"6":{"inputs":{"width":512,"height":512,"batch_size":1},"class_type":"EmptyLatentImage","_meta":{"title":"Empty Latent Image"}},"7":{"inputs":{"seed":6,"steps":20,"cfg":8,"sampler_name":"euler","scheduler":"normal","denoise":1,"model":["1",0],"positive":["4",0],"negative":["5",0],"latent_image":["6",0]},"class_type":"KSampler","_meta":{"title":"KSampler"}},"8":{"inputs":{"samples":["7",0],"vae":["1",2]},"class_type":"VAEDecode","_meta":{"title":"VAE Decode"}},"9":{"inputs":{"images":["8",0]},"class_type":"PreviewImage","_meta":{"title":"Preview Image"}}}`),
				ExtraData: []byte(`{"extra_pnginfo":{"workflow":{"last_node_id":9,"last_link_id":9,"nodes":[{"id":1,"type":"CheckpointLoaderSimple","pos":[100,130],"size":{"0":315,"1":98},"flags":{},"order":0,"mode":0,"outputs":[{"name":"MODEL","type":"MODEL","links":[3],"shape":3},{"name":"CLIP","type":"CLIP","links":[1,2],"shape":3},{"name":"VAE","type":"VAE","links":[8],"shape":3}],"properties":{"Node name for S&R":"CheckpointLoaderSimple"},"widgets_values":["v1-5-pruned-emaonly.safetensors"]},{"id":4,"type":"CLIPTextEncode","pos":[515,130],"size":{"0":400,"1":200},"flags":{},"order":2,"mode":0,"inputs":[{"name":"clip","type":"CLIP","link":1}],"outputs":[{"name":"CONDITIONING","type":"CONDITIONING","links":[4],"shape":3}],"properties":{"Node name for S&R":"CLIPTextEncode"},"widgets_values":["beautiful scenery nature glass bottle landscape, purple galaxy bottle,"]},{"id":5,"type":"CLIPTextEncode","pos":[515,460],"size":{"0":400,"1":200},"flags":{},"order":3,"mode":0,"inputs":[{"name":"clip","type":"CLIP","link":2}],"outputs":[{"name":"CONDITIONING","type":"CONDITIONING","links":[5],"shape":3}],"properties":{"Node name for S&R":"CLIPTextEncode"},"widgets_values":["text, watermark"]},{"id":6,"type":"EmptyLatentImage","pos":[100,358],"size":{"0":315,"1":106},"flags":{},"order":1,"mode":0,"outputs":[{"name":"LATENT","type":"LATENT","links":[6],"shape":3}],"properties":{"Node name for S&R":"EmptyLatentImage"},"widgets_values":[512,512,1]},{"id":8,"type":"VAEDecode","pos":[1430,130],"size":{"0":210,"1":46},"flags":{},"order":5,"mode":0,"inputs":[{"name":"samples","type":"LATENT","link":7},{"name":"vae","type":"VAE","link":8}],"outputs":[{"name":"IMAGE","type":"IMAGE","links":[9],"shape":3}],"properties":{"Node name for S&R":"VAEDecode"}},{"id":9,"type":"PreviewImage","pos":[1740,130],"size":{"0":210,"1":26},"flags":{},"order":6,"mode":0,"inputs":[{"name":"images","type":"IMAGE","link":9}],"properties":{"Node name for S&R":"PreviewImage"}},{"id":7,"type":"KSampler","pos":[1015,130],"size":{"0":315,"1":262},"flags":{},"order":4,"mode":0,"inputs":[{"name":"model","type":"MODEL","link":3},{"name":"positive","type":"CONDITIONING","link":4},{"name":"negative","type":"CONDITIONING","link":5},{"name":"latent_image","type":"LATENT","link":6}],"outputs":[{"name":"LATENT","type":"LATENT","links":[7],"shape":3}],"properties":{"Node name for S&R":"KSampler"},"widgets_values":[6,"fixed",20,8,"euler","normal",1]}],"links":[[1,1,1,4,0,"CLIP"],[2,1,1,5,0,"CLIP"],[3,1,0,7,0,"MODEL"],[4,4,0,7,1,"CONDITIONING"],[5,5,0,7,2,"CONDITIONING"],[6,6,0,7,3,"LATENT"],[7,7,0,8,0,"LATENT"],[8,1,2,8,1,"VAE"],[9,8,0,9,0,"IMAGE"]],"groups":[],"config":{},"extra":{"ds":{"scale":0.8264462809917354,"offset":[-503.8158984375007,-2.9129687499997865]}},"version":0.4,"widget_idx_map":{"7":{"seed":0,"sampler_name":4,"scheduler":5}}}},"client_id":"877ae297b3964c16b9de5be7a487def6"}`),
				OutputNodeIDs: []string{
					"9",
				},
			},
			args: args{
				p: []byte(`[92,"4c1484b8-8b71-4c92-a1e8-6179c56fe67c",{"1":{"inputs":{"ckpt_name":"v1-5-pruned-emaonly.safetensors"},"class_type":"CheckpointLoaderSimple","_meta":{"title":"Load Checkpoint"}},"4":{"inputs":{"text":"beautiful scenery nature glass bottle landscape, purple galaxy bottle,","clip":["1",1]},"class_type":"CLIPTextEncode","_meta":{"title":"CLIP Text Encode (Prompt)"}},"5":{"inputs":{"text":"text, watermark","clip":["1",1]},"class_type":"CLIPTextEncode","_meta":{"title":"CLIP Text Encode (Prompt)"}},"6":{"inputs":{"width":512,"height":512,"batch_size":1},"class_type":"EmptyLatentImage","_meta":{"title":"Empty Latent Image"}},"7":{"inputs":{"seed":6,"steps":20,"cfg":8,"sampler_name":"euler","scheduler":"normal","denoise":1,"model":["1",0],"positive":["4",0],"negative":["5",0],"latent_image":["6",0]},"class_type":"KSampler","_meta":{"title":"KSampler"}},"8":{"inputs":{"samples":["7",0],"vae":["1",2]},"class_type":"VAEDecode","_meta":{"title":"VAE Decode"}},"9":{"inputs":{"images":["8",0]},"class_type":"PreviewImage","_meta":{"title":"Preview Image"}}},{"extra_pnginfo":{"workflow":{"last_node_id":9,"last_link_id":9,"nodes":[{"id":1,"type":"CheckpointLoaderSimple","pos":[100,130],"size":{"0":315,"1":98},"flags":{},"order":0,"mode":0,"outputs":[{"name":"MODEL","type":"MODEL","links":[3],"shape":3},{"name":"CLIP","type":"CLIP","links":[1,2],"shape":3},{"name":"VAE","type":"VAE","links":[8],"shape":3}],"properties":{"Node name for S&R":"CheckpointLoaderSimple"},"widgets_values":["v1-5-pruned-emaonly.safetensors"]},{"id":4,"type":"CLIPTextEncode","pos":[515,130],"size":{"0":400,"1":200},"flags":{},"order":2,"mode":0,"inputs":[{"name":"clip","type":"CLIP","link":1}],"outputs":[{"name":"CONDITIONING","type":"CONDITIONING","links":[4],"shape":3}],"properties":{"Node name for S&R":"CLIPTextEncode"},"widgets_values":["beautiful scenery nature glass bottle landscape, purple galaxy bottle,"]},{"id":5,"type":"CLIPTextEncode","pos":[515,460],"size":{"0":400,"1":200},"flags":{},"order":3,"mode":0,"inputs":[{"name":"clip","type":"CLIP","link":2}],"outputs":[{"name":"CONDITIONING","type":"CONDITIONING","links":[5],"shape":3}],"properties":{"Node name for S&R":"CLIPTextEncode"},"widgets_values":["text, watermark"]},{"id":6,"type":"EmptyLatentImage","pos":[100,358],"size":{"0":315,"1":106},"flags":{},"order":1,"mode":0,"outputs":[{"name":"LATENT","type":"LATENT","links":[6],"shape":3}],"properties":{"Node name for S&R":"EmptyLatentImage"},"widgets_values":[512,512,1]},{"id":8,"type":"VAEDecode","pos":[1430,130],"size":{"0":210,"1":46},"flags":{},"order":5,"mode":0,"inputs":[{"name":"samples","type":"LATENT","link":7},{"name":"vae","type":"VAE","link":8}],"outputs":[{"name":"IMAGE","type":"IMAGE","links":[9],"shape":3}],"properties":{"Node name for S&R":"VAEDecode"}},{"id":9,"type":"PreviewImage","pos":[1740,130],"size":{"0":210,"1":26},"flags":{},"order":6,"mode":0,"inputs":[{"name":"images","type":"IMAGE","link":9}],"properties":{"Node name for S&R":"PreviewImage"}},{"id":7,"type":"KSampler","pos":[1015,130],"size":{"0":315,"1":262},"flags":{},"order":4,"mode":0,"inputs":[{"name":"model","type":"MODEL","link":3},{"name":"positive","type":"CONDITIONING","link":4},{"name":"negative","type":"CONDITIONING","link":5},{"name":"latent_image","type":"LATENT","link":6}],"outputs":[{"name":"LATENT","type":"LATENT","links":[7],"shape":3}],"properties":{"Node name for S&R":"KSampler"},"widgets_values":[6,"fixed",20,8,"euler","normal",1]}],"links":[[1,1,1,4,0,"CLIP"],[2,1,1,5,0,"CLIP"],[3,1,0,7,0,"MODEL"],[4,4,0,7,1,"CONDITIONING"],[5,5,0,7,2,"CONDITIONING"],[6,6,0,7,3,"LATENT"],[7,7,0,8,0,"LATENT"],[8,1,2,8,1,"VAE"],[9,8,0,9,0,"IMAGE"]],"groups":[],"config":{},"extra":{"ds":{"scale":0.8264462809917354,"offset":[-503.8158984375007,-2.9129687499997865]}},"version":0.4,"widget_idx_map":{"7":{"seed":0,"sampler_name":4,"scheduler":5}}}},"client_id":"877ae297b3964c16b9de5be7a487def6"},["9"]]`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.o

			var got PromptObj
			if err := json.Unmarshal(tt.args.p, &got); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.EqualValues(t, want, got)
		})
	}
}
