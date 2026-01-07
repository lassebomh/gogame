#version 330

/* -----------------------------------------------------------------------------
 * Constants and Defines
 * -------------------------------------------------------------------------- */

#define PLANET_RADIUS 0.5
#define ATMOSPHERE_RADIUS (PLANET_RADIUS * 1.3)

uniform float iTime;
uniform vec2 iResolution;
uniform sampler2D iChannel0;

out vec4 finalColor;
in vec2 fragTexCoord;

const vec3 ATM_CENTER = vec3(0.0);
const vec3 SUN_DIR = vec3(0.0, 0.0, 1.0);
const vec3 SUN_COLOR = vec3(1.0, 0.9, 0.9) * 4.0;
const vec3 GLARE_COL = vec3(1.0, 0.6, 0.2);
const vec3 ATM_SCATTER = 40.0 * vec3(0.2, 0.5, 1.2);
const float ATM_DENS_FALL = 4.0;

/* -----------------------------------------------------------------------------
 * Utility Functions
 * -------------------------------------------------------------------------- */

vec2 sincos(float x) {
  return vec2(sin(x), cos(x));
}

vec3 rotateY(vec3 p, float a) {
  float s = sin(a);
  float c = cos(a);
  return vec3(c * p.x + s * p.z, p.y, -s * p.x + c * p.z);
}

// Returns (tMin, tMax). If no hit, tMax < 0.
vec2 raySphereBounds(vec3 ro, vec3 rd, vec3 c, float r) {
    vec3 oc = ro - c;
    float b = dot(oc, rd);
    float c2 = dot(oc, oc) - r * r;
    float h = b * b - c2;
    if (h < 0.0) return vec2(-1.0, -1.0);
    h = sqrt(h);
    float t0 = -b - h;
    float t1 = -b + h;
    return vec2(t0, t1);
}

float sphere_intersect(vec3 ro, vec3 rd, vec3 p, float r) {
  vec3 oc = (ro - p);
  float b = dot(oc, rd);
  float c = (dot(oc, oc) - (r * r));
  float h = ((b * b) - c);
  if(h < 0.0) {
    return -1.0f;
  } else {
    return ((-b) - sqrt(h));
  }
}

float expstep(float x, float k) {
  return exp(((k * x) - k));
}

/* -----------------------------------------------------------------------------
 * Noise Functions
 * -------------------------------------------------------------------------- */

float noise_hash1_2(vec2 v) {
    vec3 v3 = vec3(v.x, v.y, v.x);
    v3 = fract(v3 * 0.1031);
    v3 = v3 + dot(v3, vec3(v3.y, v3.z, v3.x) + 33.33);
    return fract((v3.x + v3.y) * v3.z);
}

float noise_hash1_3(vec3 v) {
    vec3 v3 = fract(v * 0.1031);
    v3 = v3 + dot(v3, vec3(v3.y, v3.z, v3.x) + 33.33);
    return fract((v3.x + v3.y) * v3.z);
}

float noise_noisemix3(float a1, float b1, float c1, float d1, float a2, float b2, float c2, float d2, vec3 f) {
  vec3 u = ((f * f) * (3.0 - (2.0 * f)));
  vec3 u1 = (1.0 - u);
  return ((((((a1 * u1.x) + (b1 * u.x)) * u1.y) + (((c1 * u1.x) + (d1 * u.x)) * u.y)) * u1.z) + (((((a2 * u1.x) + (b2 * u.x)) * u1.y) + (((c2 * u1.x) + (d2 * u.x)) * u.y)) * u.z));
}

float noise_noise_value_1(vec3 p) {
  vec3 i = floor(p);
  vec3 f = fract(p);
  vec3 j = (i + 1.0);
  float a1 = noise_hash1_3(i);
  float b1 = noise_hash1_3(vec3(j.x, i.y, i.z));
  float c1 = noise_hash1_3(vec3(i.x, j.y, i.z));
  float d1 = noise_hash1_3(vec3(j.x, j.y, i.z));
  float a2 = noise_hash1_3(vec3(i.x, i.y, j.z));
  float b2 = noise_hash1_3(vec3(j.x, i.y, j.z));
  float c2 = noise_hash1_3(vec3(i.x, j.y, j.z));
  float d2 = noise_hash1_3(j);
  return noise_noisemix3(a1, b1, c1, d1, a2, b2, c2, d2, f);
}

float fbm3(vec3 p) {
  float a = 1.0;
  float t = 0.0;
  t = (t + (a * noise_noise_value_1(p)));
  a = (a * 0.5);
  p = ((2.0 * p) + 100.0);
  t = (t + (a * noise_noise_value_1(p)));
  a = (a * 0.5);
  p = ((2.0 * p) + 100.0);
  t = (t + (a * noise_noise_value_1(p)));
  a = (a * 0.5);
  p = ((2.0 * p) + 100.0);
  t = (t + (a * noise_noise_value_1(p)));
  a = (a * 0.5);
  p = ((2.0 * p) + 100.0);
  t = (t + (a * noise_noise_value_1(p)));
  a = (a * 0.5);
  p = ((2.0 * p) + 100.0);
  t = (t + (a * noise_noise_value_1(p)));
  return t;
}

/* -----------------------------------------------------------------------------
 * Atmosphere Model
 * -------------------------------------------------------------------------- */

float atmDensity(vec3 p) {
    float h = length(p - ATM_CENTER) - PLANET_RADIUS;
    float t = clamp(h / (ATMOSPHERE_RADIUS - PLANET_RADIUS), 0.0, 1.0);
    return exp(-t * ATM_DENS_FALL) * (1.0 - t);
}

vec3 atmosphereScattering(vec3 ro, vec3 rd, float tPlanetHit) {
    vec2 tAtm = raySphereBounds(ro, rd, ATM_CENTER, ATMOSPHERE_RADIUS);
    if (tAtm.y < 0.0) return vec3(0.0);

    float t0 = max(tAtm.x, 0.0);
    float t1 = tAtm.y;

    if (tPlanetHit >= 0.0) {
        t1 = min(t1, tPlanetHit);
        if (t1 <= t0) return vec3(0.0);
    }

    const int VIEW_SAMPLES = 32;
    float dt = (t1 - t0) / float(VIEW_SAMPLES);
    vec3 sum = vec3(0.0);
    float opticalDepth = 0.0;
    float mu = dot(rd, SUN_DIR);
    float phase = (1.0 + mu * mu) / (4.0 * 3.14159265 * 4.0);
    vec3 sunCol = vec3(1.0, 0.95, 0.9) * 6.0;

    for (int i = 0; i < VIEW_SAMPLES; ++i) {
        float t = t0 + (float(i) + 0.5) * dt;
        vec3 pos = ro + rd * t;
        float d = atmDensity(pos);
        opticalDepth += d * dt;
        vec3 transView = exp(-opticalDepth * ATM_SCATTER);
        float sunVis = 1.0;
        float tHitPlanet = sphere_intersect(pos, SUN_DIR, ATM_CENTER, PLANET_RADIUS);
        if (tHitPlanet > 0.0) {
            sunVis = 0.0;
        }
        sum += d * phase * transView * sunVis;
    }
    return sum * ATM_SCATTER * sunCol * (t1 - t0) / float(VIEW_SAMPLES);
}

vec3 atmosphereTransmittance(vec3 ro, vec3 rd, float tPlanetHit) {
    if (tPlanetHit < 0.0) return vec3(1.0);

    vec2 tAtm = raySphereBounds(ro, rd, ATM_CENTER, ATMOSPHERE_RADIUS);
    if (tAtm.y < 0.0) return vec3(1.0);

    float t0 = max(tAtm.x, 0.0);
    float t1 = min(tAtm.y, tPlanetHit);
    if (t1 <= t0) return vec3(1.0);

    const int TRANS_SAMPLES = 16;
    float dt = (t1 - t0) / float(TRANS_SAMPLES);
    float opticalDepth = 0.0;

    for (int i = 0; i < TRANS_SAMPLES; ++i) {
        float t = t0 + (float(i) + 0.5) * dt;
        vec3 pos = ro + rd * t;
        float d = atmDensity(pos);
        opticalDepth += d * dt;
    }
    return exp(-opticalDepth * ATM_SCATTER);
}

/* -----------------------------------------------------------------------------
 * Planet Generation
 * -------------------------------------------------------------------------- */

vec3 planet_color(vec3 p, out float height, out float landMask, out float mountainMask) {
    // texture mapping
    vec2 uv;
    // uv.x = atan(p.x,p.z)/6.2831 - 0.03*iTime;
    // uv.y = acos(p.y)/1.1416;
	  // uv.y *= 0.5;
    uv.x = atan(p.x, p.z)/6.2831 - 0.03*iTime;
    uv.y = 1.0 - (acos(p.y)/1.1416 * 0.5);
    

    vec3 col = vec3(0.2,0.3,0.4);
    vec3 te  = 0.8*texture( iChannel0, .5*uv.yx ).xyz;
         te += 0.3*texture( iChannel0, 2.5*uv.yx ).xyz;
	  col = mix( col, (vec3(0.2,0.5,0.1)*0.55 + 0.45*te + 0.5*texture( iChannel0, 15.5*uv.yx ).xyz)*0.4, smoothstep( 0.45,0.5,te.x) );

    vec3 cl = texture( iChannel0, 1.0*uv ).xxx;
	  col = mix( col, vec3(0.9), 0.75*smoothstep( 0.45,0.8,cl.x) );

    height = te.x;
    landMask = smoothstep(0.14, 0.5, te.x);
    mountainMask = smoothstep(0.5, 0.6, cl.x);
    
    vec3 baseLandCol = vec3(0.15, 0.35, 0.12); // darker, richer green
    vec3 texDetail   = 0.45*te + 0.5*texture(iChannel0, 15.5*uv.yx).xyz;
    vec3 landCol     = baseLandCol * 0.8 + texDetail * 0.2;  // stronger base color, weaker texture

    col = mix(col, landCol, smoothstep(0.4, 0.6, te.x));

    return col;
}

/* -----------------------------------------------------------------------------
 * Camera and Shading
 * -------------------------------------------------------------------------- */

vec3 perspective_camera(vec3 lookfrom, vec3 lookat, float tilt, float vfov, vec2 uv) {
  vec2 sc = sincos(tilt);
  vec3 vup = normalize(vec3(sc.x, sc.y, 0.0));
  vec3 w = normalize((lookat - lookfrom));
  vec3 u = cross(w, vup);
  vec3 v = cross(u, w);
  float wf = (1.0 / tan(((vfov * 3.14159265) / 360.0)));
  return normalize((((uv.x * u) + (uv.y * v)) + (wf * w)));
}

vec3 shade(vec3 rd, vec3 p) {
  vec3 normal = normalize(p);
  float ambient_dif = 0.03;
  vec3 dif = vec3(ambient_dif);
  float sun_dif = clamp(dot(normal, SUN_DIR) * 0.9 + 0.1, 0.0, 1.0);
  dif += SUN_COLOR * sun_dif;

  float height, landMask, mountainMask;
  float rotSpeed = 0.05;
  vec3 mate = planet_color(p, height, landMask, mountainMask) * 0.4;
  float rockSpec = pow(max(dot(reflect(-SUN_DIR, normal), -rd), 0.0), 16.0);
  mate += mountainMask * rockSpec * vec3(0.25, 0.22, 0.2);

  vec3 col = (mate * dif);
  float fres = clamp((1.0 + dot(normal, rd)), 0.0, 1.0);
  float sun_fres = (fres * clamp(dot(rd, SUN_DIR), 0.0, 1.0));
  col = (col * (1.0 - fres));
  col = (col + ((pow(sun_fres, 8.0) * vec3(0.4, 0.3, 0.1)) * 5.0));
  return col;
}

vec3 get_background(vec3 rd) {
  float sun_dif = dot(rd, SUN_DIR);
  vec3 col;
  col = (col + vec3(1.0, 0.9, 0.9) * expstep(sun_dif, 20000.0));
  col = (col + (vec3(1.0, 1.0, 0.1) * expstep(sun_dif, 10000.0)));
  // col = (col + (vec3(1.0, 0.7, 0.7) * expstep(sun_dif, 1000.0)));
  col = (col + (vec3(1.0, 0.6, 0.05) * expstep(sun_dif, 2000.0))/5.0);
  return col;
}

/* -----------------------------------------------------------------------------
 * Color Grading & Post-Processing
 * -------------------------------------------------------------------------- */

vec3 color_tonemap_aces(vec3 col) {
  return clamp(((col * ((2.51 * col) + 0.03)) / ((col * ((2.43 * col) + 0.59)) + 0.14)), 0.0, 1.0);
}

vec3 color_saturate(vec3 col, float sat) {
  float grey = dot(col, vec3(0.2125, 0.7154, 0.0721));
  return (grey + (sat * (col - grey)));
}

vec3 color_tone_1(vec3 col, float gain, float lift, float invgamma) {
  col = pow(col, vec3(invgamma));
  return (((gain - lift) * col) + lift);
}

vec3 color_gamma_correction(vec3 col) {
  return pow(col, vec3(0.454545455));
}

vec3 vignette(vec3 col, vec2 coord, float strength, float amount) {
  return (col * ((1.0 - amount) + (amount * pow(((((16.0 * coord.x) * coord.y) * (1.0 - coord.x)) * (1.0 - coord.y)), strength))));
}

vec3 dither(vec3 col, vec2 coord, float amount) {
  return clamp(col + noise_hash1_2(coord) * amount, 0.0, 1.0);
}

vec3 sun_glare(vec3 rd) {
  return GLARE_COL * pow(max(dot(SUN_DIR, rd), 0.0), 2.0) / 10.0;
}

void main() {
  vec2 fragCoord = gl_FragCoord.xy;
  vec2 res = vec2(iResolution.x, iResolution.y);
  vec2 coord = ((2.0 * (fragCoord - (res * 0.5))) / iResolution.y);
  float theta = 1.88495559 + iTime * 0.1;
  vec3 lookat = vec3(0.0, 0.0, 0.0);
  vec2 sc = 2.0 * sincos(theta);
  vec3 ro = vec3(sc.x, 0.5, sc.y);
  vec3 rd = perspective_camera(ro, lookat, 0.0, 50.0, coord);

  float t = sphere_intersect(ro, rd, vec3(0.0, 0.0, 0.0), PLANET_RADIUS);
  vec3 col = get_background(rd);
  float depth = 1.0;

  if(t >= 0.0) {
    vec3 p = (ro + (rd * t));
    col = shade(rd, p);
    depth = smoothstep(2.0, 2.0-PLANET_RADIUS, t);
  }

  vec3 atmIns = atmosphereScattering(ro, rd, t);
  vec3 trans  = atmosphereTransmittance(ro, rd, t);

  if (t >= 0.0) {
    col *= trans;
  } else {
    col = col * trans;
  }

  col += atmIns;
  col = (col + (0.2 * sun_glare(rd)));
  col = color_tonemap_aces(col);
  col = color_tone_1(col, 1.7, 0.002, 1.2);
  col = color_saturate(col, 0.9);
  col = color_gamma_correction(col);
  col = vignette(col, (fragCoord / res), 0.25, 0.7);
  col = dither(col, fragCoord, 0.01);
  finalColor = vec4(col.x, col.y, col.z, depth);
}