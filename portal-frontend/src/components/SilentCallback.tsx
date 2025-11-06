// Silent callback 페이지
// react-oidc-context의 UserManager가 자동으로 이 페이지를 iframe으로 로드하고
// OIDC provider로부터 받은 토큰을 처리합니다.
//
// 중요: 이 페이지에서는 signinSilent()를 호출하면 안 됩니다!
// UserManager가 iframe 내부에서 자동으로 토큰을 처리하고 부모 창에 전달합니다.

export function SilentCallback() {
  // react-oidc-context가 자동으로 처리하므로 아무것도 렌더링하지 않음
  // 또는 최소한의 로딩 인디케이터만 표시 (사용자는 볼 수 없음 - iframe 안에 있음)
  return null;
}
