# Orga

## 개요

- 로컬 LLM 을 이용한 파일 관리 도구
- 파일의 이름을 보고 어떤 종류의 파일인지 분류한다.
- 분류된 파일을 지정된 디렉토리로 이동시킨다. (move)

## 변수

- target_dir : 파일을 분류할 대상 디렉토리 경로 (기본값 : 현재 디렉토리)
- output_dir : 분류된 파일이 옮겨 질 디렉토리 경로 (기본값 : {target_dir}/organized)

## LLM 설정

- 파일 검색 도구 : `fd` 명령어 사용
- 문자열 검색 도구 : `rg` 명령어 사용
- llm model 기본값: 'gemma3n:e4b'

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
