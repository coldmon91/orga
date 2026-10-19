# Orga

## 개요

- 로컬 LLM 을 이용한 파일 관리 도구
- 파일의 이름을 보고 어떤 종류의 파일인지 분류한다.
- 분류된 파일을 지정된 디렉토리로 이동시킨다. (move)
- 바로 아래 하위 디렉토리는 이름으로 분류해 통째로 이동한다. 이름만으로 판단이 어려우면 내부 파일(최대 20개, 깊이 2)을 참고한다. `-R` 사용 시에는 파일 단위로 정리한다.

## 변수

- target_dir : 파일을 분류할 대상 디렉토리 경로 (기본값 : 현재 디렉토리)
- output_dir : 분류된 파일이 옮겨 질 디렉토리 경로 (기본값 : {target_dir}/orga)

## LLM 설정

- 파일 검색 도구 : `fd` 명령어 사용
- 문자열 검색 도구 : `rg` 명령어 사용
- 분류 backend: `ollama` (기본값, 로컬) 또는 `jev` (TypeSafe API, 파일·디렉토리 이름이 외부 서비스로 전송됨)
- llm model 기본값: ollama 는 'gemma3n:e4b', jev 는 'jev-latest'
- jev 사용 시 `TYPESAFE_API_KEY` 필요. 환경 변수 또는 작업 디렉토리·실행 파일 옆의 `.env` 파일에서 읽음

## 파일의 종류

- 문서 파일 (예: .docx, .pdf, .txt) : {output_dir}/documents
- 이미지 파일: {output_dir}/images
- 동영상 파일: {output_dir}/videos
- 오디오 파일: {output_dir}/audio
- 압축 파일: {output_dir}/archives
- 실행 파일: {output_dir}/executables
- 설치 파일: {output_dir}/installers
- 기타 파일: {output_dir}/others

## 설치 방법

이 도구를 사용하려면 [Go](https://go.dev/), [fd](https://github.com/sharkdp/fd), 그리고 [Ollama](https://ollama.com/)가 설치되어 있어야 합니다.

### 사전 준비

#### Go
[go.dev/doc/install](https://go.dev/doc/install)에서 다운로드하여 설치하세요.

#### fd (fd-find)
- **macOS**: `brew install fd`
- **Ubuntu/Debian**: `sudo apt install fd-find` (주의: 실행 파일 이름이 `fdfind`일 수 있습니다. 본 도구는 `fd`를 호출합니다.)
- **Windows**: `winget install sharkdp.fd`

#### Ollama
[ollama.com](https://ollama.com/)에서 다운로드하여 설치하세요. 설치 후 필요한 모델을 다운로드합니다:
```bash
ollama pull gemma3n:e4b
```

### 프로젝트 빌드

```bash
go build -o orga
```

## 사용 방법

```bash
# 기본 모델을 사용하여 현재 디렉토리 정리
./orga

# 대상 디렉토리와 모델 지정
./orga -target ~/Downloads -model gemma3:4b

# TypeSafe Jev 로 분류 (TYPESAFE_API_KEY 필요)
./orga -backend jev -target ~/Downloads

# 대상 디렉토리의 파일을 용량 순 리스트(절대 경로)로 출력
./orga -list -target ~/Downloads

# 대상 디렉토리의 파일을 용량 순 트리 구조로 출력
./orga -list -T -target ~/Downloads

# 숨김 파일을 포함하여 리스트 출력
./orga -list -H
```

## 인자 (Arguments)

- `-target`: 스캔할 디렉토리 (기본값 ".")
- `-output`: 결과 파일이 저장될 디렉토리 (기본값 "{target}/orga")
- `-backend`: 분류 backend, `ollama` 또는 `jev` (기본값 "ollama")
- `-model`: 사용할 모델명 (기본값 ollama 는 "gemma3n:e4b", jev 는 "jev-latest")
- `-R`: 하위 디렉토리를 재귀적으로 스캔 (기본값 false)
- `-list`: 대상 디렉토리의 파일을 용량별로 정렬하여 출력
- `-T`: 파일 리스트를 트리 구조로 표시 (-list와 함께 사용)
- `-H`: 리스트 모드에서 숨김 파일 표시
