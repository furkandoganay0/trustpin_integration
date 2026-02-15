# Trustpin Integration Service

Bu depo, Trustpin dış servisiyle entegrasyon yapan bir Go mikroservisini içerir. Amaç, kimlik doğrulama ve çok faktörlü doğrulama (MFA) akışlarını Trustpin aracılığıyla yönetmektir. README geliştiricilerin projeyi hızlıca çalıştırabilmesi, test edebilmesi ve geliştirebilmesi için hazırlanmıştır.

**Hedef Okuyucu**: Go geliştiricileri ve entegrasyon mühendisleri.

**Öne çıkan teknolojiler**
- Go (modül: `go 1.22`)
- PostgreSQL (opsiyonel, yoksa bellek-içi fall-back)
- Redis (nonce store için, opsiyonel)
- Docker / Docker Compose

**Hızlı Özet**
- Proje kök dizini: `cmd/server` içinde çalıştırılabilir ana uygulama bulunur.
- Konfigürasyon: `internal/config` kullanılarak ortam değişkenlerinden yüklenir.
- Trustpin entegrasyonu: `internal/adapters/trustpin`.
- Eğer `DB_DSN` veya `REDIS_ADDR` sağlanmazsa uygulama bellek-içi (in-memory) reponlarla çalışır ve bir demo kullanıcı ile başlar.

**Dosya/Dizin Haritası (kısa)**
- `cmd/server` : Uygulama giriş noktası (main).
- `internal/config` : Ortam değişkenlerini okuyan yapı.
- `internal/adapters/trustpin` : Trustpin API istemci + adapter.
- `internal/application` : Servis mantığı (AuthService, MFAService).
- `internal/infrastructure` : DB/Redis/Idempotency/jwt implementasyonları.
- `internal/transport/http` : HTTP handler'lar ve rotalar.

**Gereksinimler**
- Go 1.22
- Docker (opsiyonel, konteyner ile çalıştırmak için)
- (Opsiyonel) PostgreSQL ve Redis, eğer persistent depolama isteniyorsa.

## Hızlı Başlangıç (Lokalde)

1) Depoyu klonlayın ve köke geçin:

```bash
git clone <repo-url>
cd trustpin_integration
```

2) Ortam değişkenleri: örnek `.env` içeriği `config/.env.example` dosyasında bulunur. Geliştirme için bir kopya oluşturun:

```bash
cp config/.env.example config/.env
# Gerekliyse config/.env içindeki değerleri düzenleyin
```

3) Uygulamayı çalıştırma seçenekleri:

- Doğrudan Go ile (lokal geliştirme, bellek-içi fallback ile):

```bash
go run ./cmd/server
```

- Build edip çalıştırma:

```bash
go build -o bin/trustpin ./cmd/server
./bin/trustpin
```

Not: `DB_DSN` veya `REDIS_ADDR` boş bırakılırsa uygulama bellek-içi reponları kullanır ve `demo-user` isimli demo kullanıcı otomatik olarak eklenir.

## Docker / Docker Compose

Hazır bir docker-compose tanımı `docker/docker-compose.yml` içinde bulunmaktadır. Compose, `backend`, `db` (Postgres) ve `redis` servislerini tanımlar.

Docker ile çalıştırmak için (proje kökünden):

```bash
docker compose -f docker/docker-compose.yml up --build
```

Compose, `backend` servisi için `../config/.env` dosyasını (`docker` dizininin üstündeki `config/.env`) kullanır; bu nedenle `config/.env` dosyanızın doğru ayarlandığından emin olun.

## Ortam Değişkenleri

`internal/config.Load()` fonksiyonunda kullanılan ortam değişkenleri (varsayılanlar parantez içinde):

- `APP_ENV` : Uygulama ortamı (`dev`)
- `PORT` : HTTP portu (`8080`)
- `DB_DSN` : Postgres DSN (ör: `postgres://user:pass@host:5432/dbname?sslmode=disable`). Boşsa bellek-içi repo kullanılır.
- `REDIS_ADDR` : Redis adresi (ör: `localhost:6379`). Boşsa bellek-içi nonce store kullanılır.
- `TRUSTPIN_BASE_URL` : Trustpin temel URL'i
- `TRUSTPIN_API_KEY` : Trustpin API anahtarı
- `JWT_ISSUER` : JWT issuer
- `JWT_AUDIENCE` : JWT audience
- `JWT_PUBLIC_KEY` : Public key PEM (tek satırda `\\n` ile escape edilmiş olabilir)
- `JWT_PRIVATE_KEY` : Private key PEM
- `HTTP_TIMEOUT` : Trustpin çağrıları için timeout (ör: `5s`)
- `RETRY_MAX` : Trustpin retry maksimum deneme sayısı
- `RETRY_BACKOFF` : Retry backoff (örn: `200ms`)

Örnek `.env` içeriği: `config/.env.example` dosyasını inceleyin.

## Veri Tabanı ve Redis

- Eğer `DB_DSN` sağlanırsa `internal/infrastructure/postgres` içindeki repo implementasyonları kullanılacaktır.
- Eğer `REDIS_ADDR` sağlanırsa `internal/infrastructure/redis` içindeki nonce store kullanılacaktır.

Local geliştirmede bunları sağlamazsanız uygulama bellek-içi reponlarla çalışır — bu, hızlı geliştirme ve test için kullanışlıdır.

## API ve OpenAPI

- HTTP transport için rotalar `internal/transport/http` içinde tanımlıdır. OpenAPI tanımı `internal/transport/http/openapi.yaml` içinde bulunmaktadır.
- Projeye özel Postman koleksiyonu `docs/postman_collection.json` dosyasında bulunur.

## Trustpin Entegrasyonu

- Trustpin istemcisi `internal/adapters/trustpin/client.go` içinde yer alır; API anahtarınızı `TRUSTPIN_API_KEY` ile konfigure edin.
- `internal/adapters/trustpin/adapter.go` Trustpin çağrılarını uygulama katmanına (MFA servisleri vb.) uyarlayan adapter implementasyonudur.

## Testler

Projede varsa testleri çalıştırmak için:

```bash
go test ./...
```

Not: Bu repo geniş bir entegrasyon katmanına sahip olduğundan, gerçek Trustpin isteği yapan testler için mock/fixture veya environment config gereklidir.

## Logging ve Hata Yönetimi

- Uygulama `slog` ile JSON log üretir. `cmd/server/main.go` içinde standart çıktı yönlendirilmektedir.

## Geliştirme İpuçları
- Hızlı deneme yapmak için `config/.env` dosyanızı kullanarak `DB_DSN` ve `REDIS_ADDR`'ı boş bırakın — uygulama bellek içi modda başlatılır.
- JWT anahtarlarını `config/jwt_private.pem` ve `config/jwt_public.pem` olarak sağlamak, `config/.env` içindeki PEM alanlarını kullanmaktan daha pratiktir (ancak `config/.env` içi pem'ler de çalışır).

## Katkıda Bulunma
- Kod stili ve PR süreçleri için proje sahibine danışın. Küçük değişiklikler için fork→branch→PR akışı uygundur.

## İrtibat
- Proje sahibi: repository (GitHub) üzerinde `furkandoganay0` kullanıcı hesabı.

---
Bu README geliştiricilerin projeyi anlaması ve hızlıca çalıştırabilmesi için tasarlanmıştır. İsterseniz README'yi İngilizce'ye çevirir, daha fazla örnek ve geliştirme rehberi ekleyebilirim.
