#version 330

#define MAX_LIGHTS        4
#define LIGHT_DIRECTIONAL 0
#define LIGHT_POINT       1
#define LIGHT_SPOT        2

in vec3 fragPosition;
in vec2 fragTexCoord;
//in vec4 fragColor;
in vec3 fragNormal;

uniform vec4 colDiffuse;
uniform vec4 uvClamp;
uniform sampler2D texture0;

uniform bool fullBright;
uniform sampler2D shadowMap;
uniform vec3 playerPosition;

out vec4 finalColor;

struct Light {
  int enabled;
  int type;
  vec3 position;
  vec3 target;
  vec4 color;
  
  float cutOff;
  float outerCutOff;
  float strength;
};

uniform Light lights[MAX_LIGHTS];
uniform vec4 ambient;
uniform vec3 viewPos;

void main()
{
  // Texel color fetching from texture sampler
  
  float inView = 0;
  for (float x = -2; x <= 2; x++) {
    for (float y = -2; y <= 2; y++) {
      vec2 uv2 = ((fragPosition.xyz - playerPosition).xz + vec2(x, y)/25) / (20) + 0.5;
      uv2.x *= -1;
      inView += texture(shadowMap, uv2).g / 6;
    }
  }
  
  inView = clamp(inView, 0, 1);
  
  vec2 uv = uvClamp.xy + fragTexCoord * (uvClamp.zw - uvClamp.xy);
  vec4 texelColor = texture(texture0, uv);
  
  if (fullBright) {
    finalColor = texelColor * colDiffuse;
    return;
  }
  
  vec3 normal = normalize(fragNormal);
  vec3 viewD = normalize(viewPos - fragPosition);
  
  vec3 lightDot = vec3(0.0);
  vec3 specular = vec3(0.0);

  for (int i = 0; i < MAX_LIGHTS; i++) {
    if (lights[i].enabled == 0) continue;

    if (lights[i].type == LIGHT_DIRECTIONAL) {
    
      vec3 light = -normalize(lights[i].target - lights[i].position);
      float NdotL = max(dot(normal, light), 0.0);
      
      lightDot += lights[i].color.rgb * NdotL * lights[i].strength;

      if (NdotL > 1.0) {
        specular += pow(max(0.0, dot(viewD, reflect(-(light), normal))), 16.0) * lights[i].strength;
      }
      
    } else if (lights[i].type == LIGHT_POINT) {
      
      vec3 light = normalize(lights[i].position - fragPosition);
      float NdotL = max(dot(normal, light), 0.0);
      
      lightDot += lights[i].color.rgb * NdotL * lights[i].strength;

      if (NdotL > 1.0) {
        specular += pow(max(0.0, dot(viewD, reflect(-(light), normal))), 16.0) * lights[i].strength;
      }
      
    } else if (lights[i].type == LIGHT_SPOT) {

      vec3 light = normalize(lights[i].position - fragPosition);
      vec3 rayDir = normalize(lights[i].position - lights[i].target); 
      float theta = dot(light, rayDir); 
      float epsilon = lights[i].cutOff - lights[i].outerCutOff;
      float intensity = clamp((theta - lights[i].outerCutOff) / epsilon, 0.0, 1.0);
      
      

      float NdotL = max(dot(normal, light), 0.0);
      lightDot += (lights[i].color.rgb * NdotL * intensity * lights[i].strength) * inView;

      if (NdotL > 1.0) {
        specular += (pow(max(0.0, dot(viewD, reflect(-light, normal))), 16.0) * intensity) * inView;
      }
    }
  }
  
  finalColor = (texelColor*((colDiffuse + vec4(specular, 1.0))*vec4(lightDot, 1.0)));
  finalColor += texelColor*(ambient)*colDiffuse * clamp(inView, 0.1, 1);
  
  
  // finalColor.rgb *= inView;
}

