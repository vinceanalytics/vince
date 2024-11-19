!function(){"use strict";var p=window.location,u=window.document,l=u.currentScript,s=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var r=l&&l.getAttribute("data-include"),a=l&&l.getAttribute("data-exclude");if("pageview"===e){r=!r||r.split(",").some(n),a=a&&a.split(",").some(n);if(!r||a)return c("exclusion rule",t)}function n(e){return p.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var r={},a=(r.n=e,r.u=t&&t.u?t.u:p.href,r.d=l.getAttribute("data-domain"),r.r=u.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=r.p||{},o=(a.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);i[t]=i[t]||e}),r.p=i,new XMLHttpRequest);o.open("POST",s,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(r)),o.onreadystatechange=function(){4===o.readyState&&t&&t.callback&&t.callback({status:o.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function f(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var m=1;function a(e){var t,r,a,n;if("auxclick"!==e.type||e.button===m)return t=function(e){for(;e&&(void 0===e.tagName||!f(e)||!e.href);)e=e.parentNode;return e}(e.target),r=t&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||b<r)return!1;if(h(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(a=r)&&(n=a.split(".").pop(),g.some(function(e){return e===n}))?d(e,t,{name:"File Download",props:{url:r}}):void 0}function d(e,t,r){var a,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(a={props:r.props},plausible(r.name,a)):(a={props:r.props,callback:i},plausible(r.name,a),setTimeout(i,5e3),e.preventDefault())}u.addEventListener("click",a),u.addEventListener("auxclick",a);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=l.getAttribute("file-types"),o=l.getAttribute("add-file-types"),g=i&&i.split(",")||o&&o.split(",").concat(n)||n;function v(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{}},r=e&&e.classList;if(r)for(var a=0;a<r.length;a++){var n,i=r.item(a).match(/plausible-event-(.+)(=|--)(.+)/);i&&(n=i[1],i=i[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i)}return t}var b=3;function w(e){if("auxclick"!==e.type||e.button===m){for(var t,r,a,n,i=e.target,o=0;o<=b&&i;o++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;f(i)&&(t=i),h(i)&&(r=i),i=i.parentNode}r&&(n=v(r),t?(n.props.url=t.href,d(e,t,n)):((e={}).props=n.props,plausible(n.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}u.addEventListener("submit",function(e){var t,r=e.target,a=v(r);function n(){t||(t=!0,r.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),e={props:a.props,callback:n},plausible(a.name,e))}),u.addEventListener("click",w),u.addEventListener("auxclick",w)}();