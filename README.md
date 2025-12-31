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

## 기본 동작

- {target_dir} 내의 모든 파일을 검색한다.
- 각 파일의 이름을 LLM에 전달하여 파일의 {output_dir}을 찾는다.
- 분류된 파일을 {output_dir}의 해당하는 폴더로 이동시킨다.